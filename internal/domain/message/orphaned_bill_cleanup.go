package message

type OrphanedBillCleanup struct {
	BillObjectKeys []string `json:"billObjectKeys"`
	BucketName     string   `json:"bucketName"`
}

func (OrphanedBillCleanup) Type() string {
	return "orphaned-bill-cleanup"
}
