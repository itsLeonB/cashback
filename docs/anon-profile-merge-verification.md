# Anon-profile merge — manual verification runbook

Manual verification for the anonymous-profile physical-merge refactor
(`ProfileService.MergeAnonymousProfile`, see `internal/domain/service/profile_service.go`).
No automated test covers the merge itself — the codebase has no repository-layer
test harness — so run this against a dev/staging environment before shipping.

## Pre-deploy blocker: check for pre-existing `related_profiles` rows

This deploy deletes the old read-time consolidation logic (`ProfileService.GetAssociatedIDs`
and friends). If any `related_profiles` rows already exist in production from before this
change, their anon profiles were **never physically merged** — their debt transactions and
group-expense participation are still split across both the anon and real profile IDs, and
after this deploy there is no more read-time consolidation to paper over that split.

Before deploying, run:
```sql
select count(*) from related_profiles;
```
If nonzero, each row needs to be run through the new merge (e.g. via
`POST /profile/associate` with that row's `real_profile_id`/`anon_profile_id`, called by a
profile that's friends with both) **before** cutting over, or written up as a dedicated
one-off backfill script. This wasn't built as part of this change — it needs its own review
given it mutates financial data — so treat it as a required manual step, not something this
runbook automates.

## Scenario A: auto-merge via slug (friendship dedup + group-expense share summing)

1. Register/login as Alice (the owner). Get her `profileID` from `GET /api/v1/profile`.
2. Create an anon placeholder "Bob":
   ```
   POST /api/v1/friendships
   { "name": "Bob" }
   ```
   Note the returned friendship `id`.
3. Get Bob's slug: `GET /api/v1/friendships/{friendshipId}` → response `friend.slug`. Note
   Bob's `profileId` too (`friend.profileId`).
4. Create a draft group expense as Alice, add Bob as a participant with a nonzero share:
   ```
   POST /api/v1/group-expenses
   { "description": "test", "currency": "USD" }
   ```
   ```
   PUT /api/v1/group-expenses/{groupExpenseId}/participants
   { "participantProfileIds": ["<alice-id>", "<bob-id>"], "payerProfileId": "<alice-id>" }
   ```
5. **DB check** — confirm Bob has one `group_expense_participants` row with a `share_amount`:
   ```sql
   select id, participant_profile_id, share_amount from group_expense_participants where participant_profile_id = '<bob-id>';
   ```
6. Register a second real user "Carol" with Bob's slug attached:
   ```
   POST /api/v1/auth/register
   { "email": "carol@test.com", "password": "...", "slug": "<bob-slug>" }
   ```
7. Grab the verification token from the email (mail catcher/logs in your dev setup) and hit:
   ```
   GET /api/v1/auth/verify-registration?token=<token>
   ```
   This fires `AfterEmailVerified` → `associateBySlug` → `MergeAnonymousProfile`. Confirm it
   returns 200, not an error.
8. Get Carol's `profileID` (`GET /api/v1/profile` as Carol).
9. **DB checks** — run these:
   - Bob's profile row is gone:
     ```sql
     select * from user_profiles where id = '<bob-id>';
     ```
     → 0 rows.
   - Exactly one friendship between Alice and Carol, type `REAL`, no leftover `ANON` row:
     ```sql
     select id, profile_id1, profile_id2, type from friendships where profile_id1 in ('<alice-id>','<carol-id>') and profile_id2 in ('<alice-id>','<carol-id>');
     ```
     → exactly 1 row.
   - Carol is now the group-expense participant, share amount unchanged (no pre-existing
     Carol row to sum against yet):
     ```sql
     select participant_profile_id, share_amount from group_expense_participants where group_expense_id = '<group-expense-id>';
     ```
     → Bob's row gone, Carol's row present with the same share.
10. In the app/API, `GET /api/v1/group-expenses/{id}` as Carol — she should see herself as a
    participant. (This is the bug that motivated the change: previously Bob and Carol would've
    stayed two separate people.)

## Scenario B: collision case — sum-on-conflict + manual endpoint

Repeat steps 1–5 above with a fresh Alice/Dave(anon), but this time **before** merging, also
add the real target profile as a participant in the same group expense so a collision
actually happens:

1. Register a second real user "Erin" normally (no slug), get her `profileID`.
2. Have Alice friend-request Erin and accept, so Alice↔Erin are already real friends
   (`FriendshipRequestService` flow).
3. Add **both** Dave (anon) and Erin to the same group expense, each with a distinct share
   amount (e.g. Dave=10, Erin=5).
4. As Alice, call the manual merge endpoint directly:
   ```
   POST /api/v1/profile/associate
   { "realProfileId": "<erin-id>", "anonProfileId": "<dave-id>" }
   ```
5. **DB check** — Erin's participant row should now hold the summed share (15), Dave's row
   gone:
   ```sql
   select participant_profile_id, share_amount from group_expense_participants where group_expense_id = '<group-expense-id>';
   ```
6. **Negative check** — confirm the auth gate still holds: as a third user "Frank" who is
   *not* friends with Dave or Erin, call the same endpoint with their IDs and confirm you get
   a 403 (`ungerr.ForbiddenError`), not a silent merge.

## Priority

Run Scenario A steps 1–5 first — it's the primary flow and cheapest to catch a break in.
