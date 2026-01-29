package notification

import (
	"encoding/json"
	"sync"

	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/ungerr"
)

type TitleResolver interface {
	Type() string
	ResolveTitle(notification entity.Notification) (string, error)
}

var (
	once        sync.Once
	resolverMap map[string]TitleResolver
)

func ResolveTitle(n entity.Notification) (string, error) {
	resolver, err := getResolverByType(n.Type)
	if err != nil {
		return "", err
	}
	return resolver.ResolveTitle(n)
}

func getResolverByType(t string) (TitleResolver, error) {
	once.Do(func() { resolverMap = constructResolverMap() })
	resolver, exists := resolverMap[t]
	if !exists {
		return nil, ungerr.Unknownf("unknown notification type: %s", t)
	}
	return resolver, nil
}

func constructResolverMap() map[string]TitleResolver {
	resolvers := []TitleResolver{
		debtCreatedResolver{},
		expenseConfirmedResolver{},
		friendRequestReceivedResolver{},
		friendshipCreatedResolver{},
	}

	resolverMap := make(map[string]TitleResolver, len(resolvers))
	for _, resolver := range resolvers {
		resolverMap[resolver.Type()] = resolver
	}

	return resolverMap
}

func unmarshal[T any](data []byte) (T, error) {
	var zero T
	if err := json.Unmarshal(data, &zero); err != nil {
		return zero, ungerr.Wrapf(err, "error unmarshaling data to %T", zero)
	}
	return zero, nil
}
