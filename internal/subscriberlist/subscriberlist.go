package subscriberlist

import (
	"context"
	"log"

	"github.com/Exayn/go-listmonk"
	"github.com/thoas/go-funk"
)

type SubscriberList struct {
	ManagedLists          []uint
	OneTimeSubscribeLists []uint
	Lmc                   *listmonk.Client
	Ctx                   context.Context
	Memberidfield         string
	MemberSubscribers     map[int]Subscriber
	NonMemberSubscribers  map[string]Subscriber
}

func (sl *SubscriberList) LoadSubscribers() (c int, err error) {
	lmsl, err := sl.Lmc.NewGetSubscribersService().PerPage("all").Do(sl.Ctx)
	if err != nil {
		return
	}
	c = len(lmsl)
	var subscribers []Subscriber
	for _, s := range lmsl {
		subscribers = append(subscribers, NewSubscriberFromListmonk(s, sl.Memberidfield))
	}

	for _, s := range subscribers {
		if s.MemberId != nil {
			sl.MemberSubscribers[*s.MemberId] = s
		} else {
			sl.NonMemberSubscribers[s.Email] = s
		}
	}

	return
}

func NewSubscriberList(
	lmc *listmonk.Client,
	ctx context.Context,
	managedLists []uint,
	oneTimeSubscribeLists []uint,
	memberidfield string,
) *SubscriberList {

	if len(funk.Join(managedLists, oneTimeSubscribeLists, funk.InnerJoin).([]uint)) > 0 {
		log.Fatalln("Lists cannot be in both ManagedLists and OneTimeSubscribeLists")
	}

	return &SubscriberList{
		ManagedLists:          managedLists,
		OneTimeSubscribeLists: oneTimeSubscribeLists,
		Lmc:                   lmc,
		Ctx:                   ctx,
		Memberidfield:         memberidfield,
		MemberSubscribers:     make(map[int]Subscriber),
		NonMemberSubscribers:  make(map[string]Subscriber),
	}
}
