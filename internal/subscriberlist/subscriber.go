package subscriberlist

import (
	"strconv"
	"strings"

	"github.com/Exayn/go-listmonk"
	"github.com/thoas/go-funk"
)

type Subscriber struct {
	Id            uint
	Email         string
	Name          string
	Lists         []uint
	RawAttributes map[string]interface{}
	MemberId      *int
}

func (s Subscriber) Attributes(memberidfield string) (attrs map[string]interface{}) {
	attrs = make(map[string]interface{})
	for k, v := range s.RawAttributes {
		attrs[k] = v
	}
	if s.MemberId != nil {
		attrs[memberidfield] = *s.MemberId
	}
	return
}

func NewSubscriberFromListmonk(l *listmonk.Subscriber, memberidfield string) (s Subscriber) {
	s.Id = l.Id
	s.Email = strings.ToLower(l.Email)
	s.Name = l.Name
	s.Lists = funk.Map(l.Lists, func(ml listmonk.SubscriberList) uint {
		return ml.Id
	}).([]uint)
	s.RawAttributes = l.Attributes
	if memberIdRaw, ok := l.Attributes[memberidfield]; ok {
		//goland:noinspection GoShadowedVar
		if memberIdInt, ok := memberIdRaw.(int); ok {
			s.MemberId = &memberIdInt
		} else if memberIdFloat, ok := memberIdRaw.(float64); ok {
			s.MemberId = func() *int { v := int(memberIdFloat); return &v }()
		} else if meberIdStr, ok := memberIdRaw.(string); ok {
			if memberId, err := strconv.Atoi(meberIdStr); err == nil {
				s.MemberId = &memberId
			}
		}
		delete(s.RawAttributes, memberidfield)
	}
	return
}
