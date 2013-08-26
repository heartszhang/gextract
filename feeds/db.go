package feeds

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"time"
)

var (
	server  = "localhost"
	db      = "feeds"
	session *mgo.Session
)

func clone_session() *mgo.Session {
	if session == nil {
		var err error
		session, err = mgo.Dial(server)
		try_panic(err)
	}
	return session.Clone()
}

func TerminateMongo() {
	if session != nil {
		sess := session
		session = nil
		sess.Close()
	}
}
func DoInSession(collection string, act func(*mgo.Collection)) {
	sess := clone_session()
	defer sess.Close()

	c := sess.DB(db).C(collection)
	act(c)
}

func InsertChannel(c Channel) int {
	rtn := 0
	DoInSession("channels", func(coll *mgo.Collection) {
		cond := bson.M{"link": c.Href}
		ci, err := coll.Upsert(cond, &c)
		if ci.Updated > 0 {
			log.Println("insert dup channel", c.Title, c.Href)
		}
		rtn = ci.Updated
		try_panic(err)
	})
	return rtn
}

func insert_entry(coll *mgo.Collection, e Entry) {
	cond := bson.M{
		"$or": []bson.M{
			{"link": e.Link},
			{"guid": e.Guid},
			{"title": e.Title}}}
	ci, err := coll.Upsert(cond, &e)
	if ci.Updated > 0 {
		log.Println("insert dup entry", e.Link, e.Title, e.Guid)
	}
	try_panic(err)
}

func InsertEntry(e Entry) {
	DoInSession("entries", func(coll *mgo.Collection) {
		insert_entry(coll, e)
	})
}

func SaveEntries(channel Channel, entries []Entry) {
	InsertEntries(entries)
	InsertChannel(channel)
}
func InsertEntries(entries []Entry) {
	DoInSession("entries", func(coll *mgo.Collection) {
		for _, e := range entries {
			insert_entry(coll, e)
		}
	})
}

func TopNEntries(skip, limit int) []Entry {
	set := []Entry{}
	DoInSession("entries", func(coll *mgo.Collection) {
		err := coll.Find(bson.M{"read": false}).Sort("-created").Skip(skip).Limit(limit).All(&set)
		try_panic(err)
	})
	return set
}

func TopNEntriesByCategory(category string, skip, limit int) []Entry {
	set := []Entry{}
	DoInSession("entries", func(coll *mgo.Collection) {
		err := coll.Find(bson.M{"category": category, "read": false}).Sort("-pubdate").Skip(skip).Limit(limit).All(&set)
		try_panic(err)
	})
	return set
}

func EntriesContentUnready() []Entry {
	set := []Entry{}
	DoInSession("entries", func(coll *mgo.Collection) {
		err := coll.Find(bson.M{"content_status": content_status_unresolved, "read": false}).All(&set)
		if err != nil {
			log.Println(err)
		}
	})
	return set

}

func TopNEntriesByChannel(cid bson.ObjectId, skip, limit int) []Entry {
	set := []Entry{}
	DoInSession("entries", func(coll *mgo.Collection) {
		err := coll.Find(bson.M{"cid": cid, "read": false}).Sort("-pubdate").Skip(skip).Limit(limit).All(&set)
		try_panic(err)
	})
	return set
}

func RefreshChannels() []Channel {
	set := []Channel{}
	DoInSession("channels", func(coll *mgo.Collection) {
		err := coll.Find(bson.M{"refresh": bson.M{"$gt": time.Now()}}).All(&set)
		try_panic(err)
	})
	return set
}

func RefreshCategoryChannels(category string) []Channel {
	set := []Channel{}
	DoInSession("channels", func(coll *mgo.Collection) {
		err := coll.Find(bson.M{"refresh": bson.M{"$gt": time.Now()}, "category": category}).All(&set)
		try_panic(err)
	})
	return set
}

func TouchChannel(href string, ttl int) {
	nextt := time.Now().Add(time.Duration(ttl) * time.Minute)
	DoInSession("channels", func(coll *mgo.Collection) {
		_, err := coll.UpdateAll(bson.M{"link": href},
			bson.M{"$inc": bson.M{"refresh": nextt}})
		try_panic(err)
	})
}

/*
	content_status_ready      // content has been stored in a file
	content_status_failed     // fetch content failed
	content_status_unavail    // fetch content ok, buy extracted-body is low quality
	content_status_summary    // content stored in Entry.Summary
	content_status_content    // content stored in Entry.Content
*/
func EntryUpdateContent(tf, link string) {
	DoInSession("entries", func(coll *mgo.Collection) {
		cs := content_status_failed
		if len(tf) > 0 {
			cs = content_status_ready
		}
		coll.Update(bson.M{"link": link}, bson.M{"$inc": bson.M{"content_status": cs, "content_path": tf}})
	})
}

func MarkReadCatetoryaDayBefore(category string) {
	deadline := time.Now().AddDate(0, 0, -1)
	DoInSession("entries", func(coll *mgo.Collection) {
		_, err := coll.UpdateAll(bson.M{"pubdate": bson.M{"$lt": deadline}, "category": category},
			bson.M{"$inc": bson.M{"read": true}})
		try_panic(err)
	})
}

func MarkReadChannelaDayBefore(cid string) {
	deadline := time.Now().AddDate(0, 0, -1)
	DoInSession("entries", func(coll *mgo.Collection) {
		_, err := coll.UpdateAll(bson.M{"pubdate": bson.M{"$lt": deadline}, "cid": cid},
			bson.M{"$inc": bson.M{"read": true}})
		try_panic(err)
	})
}

func MarkReadEntry(link string) {
	DoInSession("entries", func(coll *mgo.Collection) {
		err := coll.Update(bson.M{"link": link},
			bson.M{"$inc": bson.M{"read": true}})
		try_panic(err)
	})

}
