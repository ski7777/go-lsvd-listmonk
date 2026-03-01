package memberlist

import (
	"strings"
	"time"

	"github.com/thoas/go-funk"
)

type Member struct {
	MemberId        *int       `lsvd.field:"MitglNr" lsvd.must:"true"`
	Title           *string    `lsvd.field:"AnredeTitel"`
	FirstName       *string    `lsvd.field:"Vorname"`
	LastName        *string    `lsvd.field:"Nachname"`
	Company         *string    `lsvd.field:"Firma1"`
	Mail            *string    `lsvd.field:"Email" lsvd.lower:"true"`
	ResignationDate *time.Time `lsvd.field:"Austritt" lsvd.timetype:"dateonly"`
}

func (m Member) FullName() string {
	if m.Company != nil && *m.Company != "" {
		return *m.Company
	}
	return strings.Join(
		funk.Map(
			funk.Filter(
				[]*string{m.Title, m.FirstName, m.LastName},
				func(s *string) bool {
					return s != nil && *s != ""
				},
			).([]*string),
			func(s *string) string {
				return *s
			},
		).([]string),
		" ",
	)
}

func (m Member) Valid() bool {
	return m.ResignationDate == nil || m.ResignationDate.After(time.Now())
}
