package memberlist

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/structtag"
	"github.com/xuri/excelize/v2"
)

func Load(filecontent []byte) (members map[int]Member, err error) {
	f, err := excelize.OpenReader(bytes.NewReader(filecontent),
		excelize.Options{
			ShortDatePattern: "yyyy-mm-dd",
		})
	if err != nil {
		return
	}
	defer func() {
		if errclose := f.Close(); errclose != nil && err == nil {
			err = errclose
		}
	}()
	var rows *excelize.Rows
	if sheets := f.GetSheetList(); len(sheets) != 1 {
		err = errors.New("unexpected number of sheets")
		return
	} else if rows, err = f.Rows(sheets[0]); err != nil {
		return
	}
	if !rows.Next() {
		err = errors.New("no rows found")
		return
	}
	titles, err := rows.Columns()
	if err != nil {
		return
	}
	titlemap := make(map[string]int)
	for i, title := range titles {
		titlemap[title] = i
	}
	type FieldInfo struct {
		ColIndex int
		Type     reflect.Type
		Must     bool
		Lower    bool
		TimeType string
	}
	structfieldsfieldmap := make(map[string]FieldInfo) // maps struct field name to column index
	structfields := reflect.VisibleFields(reflect.TypeOf(Member{}))
	for _, structfield := range structfields {
		var tags *structtag.Tags
		tags, err = structtag.Parse(string(structfield.Tag))
		if err != nil {
			return
		}
		tagmap := make(map[string]*structtag.Tag)
		for _, tag := range tags.Tags() {
			if strings.HasPrefix(tag.Key, "lsvd.") {
				tagmap[tag.Key] = tag
			}
		}
		if lsvdfield, ok := tagmap["lsvd.field"]; ok {
			//goland:noinspection GoShadowedVar
			if colindex, ok := titlemap[lsvdfield.Name]; ok {
				structfieldsfieldmap[structfield.Name] = FieldInfo{
					ColIndex: colindex,
					Type:     structfield.Type,
					Must: func() bool {
						//goland:noinspection GoShadowedVar
						if musttag, ok := tagmap["lsvd.must"]; ok {
							must, errparse := strconv.ParseBool(musttag.Name)
							if errparse != nil {
								err = fmt.Errorf("failed to parse lsvd.must tag value %q: %w", musttag.Name, errparse)
								return false
							}
							return must
						}
						return false
					}(),
					Lower: func() bool {
						if structfield.Type != reflect.TypeOf((*string)(nil)) {
							return false
						}
						//goland:noinspection GoShadowedVar
						if lowertag, ok := tagmap["lsvd.lower"]; ok {
							must, errparse := strconv.ParseBool(lowertag.Name)
							if errparse != nil {
								err = fmt.Errorf("failed to parse lsvd.must tag value %q: %w", lowertag.Name, errparse)
								return false
							}
							return must
						}
						return false
					}(),
					TimeType: func() string {
						//goland:noinspection GoShadowedVar
						if tttag, ok := tagmap["lsvd.timetype"]; ok {
							return tttag.Name
						}
						return ""
					}(),
				}
			} else {
				err = fmt.Errorf("no column for lsvd.field %s", lsvdfield.Name)
				return
			}
		} else {
			err = fmt.Errorf("no lsvd.field for struct field %s", structfield.Name)
			return
		}
	}
	members = make(map[int]Member)
	mail := make(map[string]struct{})
	for rows.Next() {
		var cols []string
		cols, err = rows.Columns()
		if err != nil {
			return
		}
		member := Member{}
		reflectedMember := reflect.ValueOf(&member)
		for structfieldname, fieldinfo := range structfieldsfieldmap {
			rawval := strings.TrimSpace(cols[fieldinfo.ColIndex])
			if len(rawval) > 0 {
				switch fieldinfo.Type {
				case reflect.TypeOf((*string)(nil)):
					reflectedMember.Elem().FieldByName(structfieldname).Set(reflect.ValueOf(&rawval))
				case reflect.TypeOf((*int)(nil)):
					var rawint int
					rawint, err = strconv.Atoi(rawval)
					if err != nil {
						err = fmt.Errorf("failed to parse int value %q: %w", rawval, err)
						return
					}
					reflectedMember.Elem().FieldByName(structfieldname).Set(reflect.ValueOf(&rawint))
				case reflect.TypeOf((*time.Time)(nil)):
					var parsedTime time.Time
					switch fieldinfo.TimeType {
					case "dateonly":
						parsedTime, err = time.Parse(time.DateOnly, rawval)
						if err != nil {
							err = fmt.Errorf("failed to parse dateonly value %q: %w", rawval, err)
							return
						}
					default:
						err = fmt.Errorf("unsupported timetype %q for field %s", fieldinfo.TimeType, structfieldname)
						return
					}
					reflectedMember.Elem().FieldByName(structfieldname).Set(reflect.ValueOf(&parsedTime))
				default:
					err = fmt.Errorf("unsupported field type %s", fieldinfo.Type.String())
					return
				}
			}
			if fieldinfo.Must {
				fieldval := reflectedMember.Elem().FieldByName(structfieldname)
				if fieldval.IsNil() {
					err = fmt.Errorf("missing required field %s", structfieldname)
					return
				}
			}
			if fieldinfo.Lower {
				fieldval := reflectedMember.Elem().FieldByName(structfieldname)
				if !fieldval.IsNil() {
					lowered := strings.ToLower(*fieldval.Interface().(*string))
					fieldval.Set(reflect.ValueOf(&lowered))
				}
			}
		}
		if !member.Valid() {
			log.Printf("Member %s (%d) is invalid (resigned). Ommitting.\n", member.FullName(), *member.MemberId)
			continue
		}
		if member.Mail != nil {
			if _, ok := mail[*member.Mail]; ok {
				log.Println("Duplicate mail address", *member.Mail, ". Ommitting member ID", *member.MemberId)
				continue
			}
			mail[*member.Mail] = struct{}{}
		}
		members[*member.MemberId] = member
	}
	return
}
