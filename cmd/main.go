package main

import (
	"context"
	"log"

	"github.com/Exayn/go-listmonk"
	"github.com/ski7777/go-lsvd-listmonk/internal/config"
	"github.com/ski7777/go-lsvd-listmonk/internal/memberlist"
	"github.com/ski7777/go-lsvd-listmonk/internal/sharepointfile"
	"github.com/ski7777/go-lsvd-listmonk/internal/subscriberlist"
	"github.com/ski7777/go-lsvd-listmonk/internal/syncer"
)

func main() {
	c, err := config.NewConfigFromFile("/config/config.json")
	if err != nil {
		log.Fatalln(err)
	}

	lm := subscriberlist.NewSubscriberList(
		listmonk.NewClient(c.Listmonk.Url, &c.Listmonk.Credentials.Username, &c.Listmonk.Credentials.Password),
		context.Background(),
		c.Listmonk.Lists.Managed,
		c.Listmonk.Lists.OneTimeSubscribe,
		c.Listmonk.MemberIdField,
	)
	sc, err := lm.LoadSubscribers()
	if err != nil {
		return
	}

	log.Println("Subscribers:", sc)
	log.Println("Subscribers with member ID:", len(lm.MemberSubscribers))
	log.Println("Subscribers without member ID:", len(lm.NonMemberSubscribers))

	excelfile, err := sharepointfile.GetFile(c.Sharepoint.Credentials.TenantId, c.Sharepoint.Credentials.ClientId, c.Sharepoint.Credentials.ClientSecret, c.Sharepoint.DriveId, c.Sharepoint.FolderItemId)
	if err != nil {
		log.Fatalln(err)
	}

	members, err := memberlist.Load(excelfile)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Members loaded:", len(members))

	err = syncer.Sync(lm, members)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Sync completed successfully")
}
