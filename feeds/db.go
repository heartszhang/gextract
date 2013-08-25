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
	//	sess, err := mgo.Dial(server)
	//	try_panic(err)
	sess := clone_session()
	defer sess.Close()

	c := sess.DB(db).C(collection)
	act(c)
}

func InsertChannel(c Channel) {
	DoInSession("channels", func(coll *mgo.Collection) {
		cond := bson.M{"link": c.Href}
		_, err := coll.Upsert(cond, &c)
		try_panic(err)
	})
}

func InsertEntry(e Entry) {
	DoInSession("entries", func(coll *mgo.Collection) {
		cond := bson.M{
			"$or": []bson.M{
				{"link": e.Link},
				{"guid": e.Guid},
				{"title": e.Title},
				{"digest": e.Digest}}}
		_, err := coll.Upsert(cond, &e)
		try_panic(err)
	})
}

func InsertEntries(entries []Entry) {
	DoInSession("entries", func(coll *mgo.Collection) {
		for _, e := range entries {
			cond := bson.M{"$or": []bson.M{{"link": e.Link},
				{"guid": e.Guid},
				{"title": e.Title},
				{"digest": e.Digest}}}
			ci, err := coll.Upsert(cond, &e)
			if ci.Updated > 0 {
				log.Println("insert dup entry", e.Link, e.Title, e.Digest)
			}
			try_panic(err)
		}
	})
}

func TopNEntries(skip, limit int) []Entry {
	set := []Entry{}
	DoInSession("entries", func(coll *mgo.Collection) {
		err := coll.Find(nil).Sort("-pubdate").Skip(skip).Limit(limit).All(&set)
		try_panic(err)
	})
	return set
}

func TopNEntriesByCategory(category string, skip, limit int) []Entry {
	set := []Entry{}
	DoInSession("entries", func(coll *mgo.Collection) {
		err := coll.Find(bson.M{"category": category}).Sort("-pubdate").Skip(skip).Limit(limit).All(&set)
		try_panic(err)
	})
	return set
}

func TopNEntriesByChannel(cid bson.ObjectId, skip, limit int) []Entry {
	set := []Entry{}
	DoInSession("entries", func(coll *mgo.Collection) {
		err := coll.Find(bson.M{"cid": cid}).Sort("-pubdate").Skip(skip).Limit(limit).All(&set)
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

func MarkReadCatetoryaDayBefore(category string) {
	deadline := time.Now().AddDate(0, 0, -1)
	DoInSession("entries", func(coll *mgo.Collection) {
		_, err := coll.UpdateAll(bson.M{"pubdate": bson.M{"$lt": deadline}, "category": category},
			bson.M{"$inc": bson.M{"read": true}})
		try_panic(err)
	})
}

func MarkReadChannelaDayBefore(cid bson.ObjectId) {
	deadline := time.Now().AddDate(0, 0, -1)
	DoInSession("entries", func(coll *mgo.Collection) {
		_, err := coll.UpdateAll(bson.M{"pubdate": bson.M{"$lt": deadline}, "cid": cid},
			bson.M{"$inc": bson.M{"read": true}})
		try_panic(err)
	})
}

func MarkReadEntry(id bson.ObjectId) {
	DoInSession("entries", func(coll *mgo.Collection) {
		err := coll.Update(bson.M{"_id": id},
			bson.M{"$inc": bson.M{"read": true}})
		try_panic(err)
	})

}
