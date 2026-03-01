package syncer

import (
	"errors"
	"fmt"
	"log"

	"github.com/Exayn/go-listmonk"
	"github.com/ski7777/go-lsvd-listmonk/internal/memberlist"
	"github.com/ski7777/go-lsvd-listmonk/internal/subscriberlist"
	"github.com/thoas/go-funk"
)

func Sync(lm *subscriberlist.SubscriberList, members map[int]memberlist.Member) (err error) {
	//Remove non-members
	for _, mid := range funk.LeftJoinInt(funk.Keys(lm.MemberSubscribers).([]int), funk.Keys(members).([]int)) {
		log.Printf("Member ID %d in subscribers but not in members\n", mid)
		m := lm.MemberSubscribers[mid]
		delete(lm.MemberSubscribers, mid)
		m.MemberId = nil
		for _, lid := range funk.Join(m.Lists, lm.ManagedLists, funk.InnerJoin).([]uint) {
			log.Printf(" - Subscriber %s is subscribed to list %d. Unsubscribing\n", m.Email, lid)
		}
		m.Lists = funk.Join(m.Lists, lm.ManagedLists, funk.LeftJoin).([]uint)
		lm.NonMemberSubscribers[m.Email] = m
		_, err = lm.Lmc.NewUpdateSubscriberService().
			Id(m.Id).
			Email(m.Email).
			Attributes(m.Attributes(lm.Memberidfield)).
			PreconfirmSubscriptions(true).
			Name(m.Name).
			ListIds(m.Lists).
			Do(lm.Ctx)
		if err != nil {
			return
		}
	}

	//Add missing members
	for _, mid := range funk.RightJoinInt(funk.Keys(lm.MemberSubscribers).([]int), funk.Keys(members).([]int)) {
		m := members[mid]
		if m.Mail == nil {
			continue
		}
		log.Println("Member ID", mid, "in members but not in subscribers")
		if s, ok := lm.NonMemberSubscribers[*m.Mail]; ok {
			log.Println(" - Subscriber", s.Email, "matches member", *m.MemberId)
			delete(lm.NonMemberSubscribers, s.Email)
			s.MemberId = &mid
			lm.MemberSubscribers[mid] = s
			_, err = lm.Lmc.NewUpdateSubscriberService().
				Id(s.Id).
				Email(s.Email).
				Attributes(s.Attributes(lm.Memberidfield)).
				PreconfirmSubscriptions(true).
				ListIds(s.Lists).
				Do(lm.Ctx)
			if err != nil {
				return
			}
		} else {
			log.Println(" - Creating new subscriber for member", *m.MemberId, *m.Mail)
			ns := subscriberlist.Subscriber{
				Id:       0, // to be filled after creation,
				Email:    *m.Mail,
				Lists:    funk.UniqUInt(append(lm.ManagedLists, lm.OneTimeSubscribeLists...)),
				MemberId: m.MemberId,
				Name:     m.FullName(),
			}
			var sn *listmonk.Subscriber
			sn, err = lm.Lmc.NewCreateSubscriberService().
				Email(ns.Email).
				Name(ns.Name).
				Attributes(ns.Attributes(lm.Memberidfield)).
				PreconfirmSubscriptions(true).
				ListIds(ns.Lists).
				Do(lm.Ctx)
			if err != nil {
				return
			}
			ns.Id = sn.Id
			lm.MemberSubscribers[mid] = ns
		}
	}

	if len(lm.MemberSubscribers) != len(
		funk.Filter(
			funk.Keys(members).([]int),
			func(mid int) bool {
				return members[mid].Mail != nil
			},
		).([]int),
	) {
		err = errors.New("mismatch between members and lm.MemberSubscribers after sync")
		return
	}

	for mail, s := range lm.NonMemberSubscribers {
		unsub := funk.Join(s.Lists, lm.ManagedLists, funk.InnerJoin).([]uint)
		for _, lid := range unsub {
			log.Printf("Non-member Subscriber %s is subscribed to managed list %d. Unsubscribing.\n", mail, lid)
		}
		s.Lists = funk.Join(s.Lists, lm.ManagedLists, funk.LeftJoin).([]uint)
		lm.NonMemberSubscribers[mail] = s
		if len(unsub) > 0 {
			_, err = lm.Lmc.NewUpdateSubscriberService().
				Id(s.Id).
				Email(s.Email).
				Attributes(s.Attributes(lm.Memberidfield)).
				PreconfirmSubscriptions(true).
				Name(s.Name).
				ListIds(s.Lists).
				Do(lm.Ctx)
			if err != nil {
				return
			}
		}
	}

	for mid, s := range lm.MemberSubscribers {
		m := members[mid]
		if *s.MemberId != *m.MemberId {
			err = fmt.Errorf("Mismatch between subscriber member ID %d and member member ID %d after sync\n", *s.MemberId, *m.MemberId)
			return
		}
		action := false
		if s.Email != *m.Mail {
			log.Println("Member ID", mid, "email changed from", *m.Mail, "to", s.Email)
			s.Email = *m.Mail
			action = true
		}
		if fn := m.FullName(); s.Name != fn {
			log.Println("Member ID", mid, "name changed from", s.Name, "to", fn)
			s.Name = fn
			action = true
		}
		for _, lid := range funk.Join(s.Lists, lm.ManagedLists, funk.RightJoin).([]uint) {
			log.Printf("Member ID %d subscriber %s is not subscribed to managed list %d. Subscribing.\n", mid, s.Email, lid)
			action = true
			s.Lists = append(s.Lists, lid)
		}
		if action {
			lm.MemberSubscribers[mid] = s
			_, err = lm.Lmc.NewUpdateSubscriberService().
				Id(s.Id).
				Email(s.Email).
				Attributes(s.Attributes(lm.Memberidfield)).
				PreconfirmSubscriptions(true).
				Name(s.Name).
				ListIds(s.Lists).
				Do(lm.Ctx)
			if err != nil {
				return
			}
		}
	}
	return
}
