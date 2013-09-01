package feeds

import (
	. "github.com/heartszhang/gextract/feeds/meta"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

type FeedOperator interface {
	Save(feeds []Feed) error
	Upsert(f *Feed) error
	Find(uri string) (*Feed, error)
	TimeoutFeeds() ([]Feed, error)
	AllFeeds() ([]Feed, error)
	Touch(uri string, ttl int) error
	Remove(link string) error
	Disable(link string, dis bool) error
	Update(f *Feed) error
}

func NewFeedOperator() FeedOperator {
	return &op_feed{}
}

type op_feed struct {
}

func (op_feed) Save(feeds []Feed) error {
	return do_in_session("feeds", func(coll *mgo.Collection) error {
		for _, f := range feeds {
			_, err := coll.Upsert(bson.M{"link": f.Link}, bson.M{"$setOnInsert": f})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (op_feed) Upsert(f *Feed) error {
	return do_in_session("feeds", func(coll *mgo.Collection) error {
		_, err := coll.Upsert(bson.M{"link": f.Link}, bson.M{"$setOnInsert": f})
		return err
	})
}

func (op_feed) Remove(link string) error {
	return do_in_session("feeds", func(coll *mgo.Collection) error {
		return coll.Remove(bson.M{"link": link})
	})
}

func (op_feed) Disable(link string, dis bool) error {
	return do_in_session("feeds", func(coll *mgo.Collection) error {
		return coll.Update(bson.M{"link": link}, bson.M{"$set": bson.M{"disabled": dis}})
	})
}

func (op_feed) Update(f *Feed) error {
	return do_in_session("feeds", func(coll *mgo.Collection) error {
		return coll.Update(bson.M{"link": f.Link},
			bson.M{
				"$set":      bson.M{"ttl": f.TTL},
				"$addToSet": bson.M{"category": bson.M{"$each": f.Category}}})
	})
}

func (op_feed) Find(uri string) (*Feed, error) {
	rtn := new(Feed)
	err := do_in_session("feeds", func(coll *mgo.Collection) error {
		err := coll.Find(bson.M{"link": uri}).One(rtn)
		return err
	})
	return rtn, err
}

func (op_feed) AllFeeds() (feds []Feed, err error) {
	feds = make([]Feed, 0)
	err = do_in_session("feeds", func(coll *mgo.Collection) error {
		return coll.Find(bson.M{"disabled": false}).All(&feds)
		//		return coll.Find(bson.M{"$nor": []bson.M{bson.M{"disabled": bson.M{"$exist": true}},
		//			bson.M{"disabled": true}}}).All(&feds)
	})
	return
}
func (op_feed) TimeoutFeeds() ([]Feed, error) {
	rtn := make([]Feed, 0)
	err := do_in_session("feeds", func(coll *mgo.Collection) error {
		return coll.Find(bson.M{"disabled": false, "refresh": bson.M{"$lt": time.Now()}}).All(&rtn)
	})
	return rtn, err
}

func (op_feed) Touch(uri string, ttl int) error {
	dl := time.Now().Add(time.Duration(ttl) * time.Minute)
	return do_in_session("feeds", func(coll *mgo.Collection) error {
		return coll.Update(bson.M{"link": uri}, bson.M{"$set": bson.M{"refresh": dl}})
	})
}

type EntryOperator interface {
	Save([]Entry) error
	SaveOne(Entry) error
	TopN(skip, limit int) ([]Entry, error)
	TopNByCategory(skip, limit int, category string) ([]Entry, error)
	TopNByFeed(skip, limit int, feed string) ([]Entry, error)
	MarkRead(link string, readed bool) error
	SetContent(link string, filepath string, words int, imgs []string) error
}

func NewEntryOperator() EntryOperator {
	return new(op_entry)
}

type op_entry struct {
}

func (op_entry) MarkRead(link string, readed bool) error {
	return do_in_session("entries", func(coll *mgo.Collection) error {
		return coll.Update(bson.M{"link": link}, bson.M{"$set": bson.M{"readed": readed}})
	})
}

func (op_entry) Save(entries []Entry) error {
	return do_in_session("entries", func(coll *mgo.Collection) error {
		for _, entry := range entries {
			err := insert_entry(coll, entry)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (op_entry) SaveOne(entry Entry) error {
	return do_in_session("entries", func(coll *mgo.Collection) error {
		return insert_entry(coll, entry)
	})
}

func (op_entry) TopN(skip, limit int) ([]Entry, error) {
	rtn := make([]Entry, 0)
	err := do_in_session("entries", func(coll *mgo.Collection) error {
		return coll.Find(bson.M{"readed": false}).Sort("-created").Skip(skip).Limit(limit).All(&rtn)
	})
	return rtn, err
}

func (op_entry) TopNByFeed(skip, limit int, feed string) ([]Entry, error) {
	rtn := make([]Entry, 0)
	err := do_in_session("entries", func(coll *mgo.Collection) error {
		return coll.Find(bson.M{"readed": false, "feed": feed}).Sort("-created").Skip(skip).Limit(limit).All(&rtn)
	})
	return rtn, err
}

func (op_entry) TopNByCategory(skip, limit int, category string) ([]Entry, error) {
	rtn := make([]Entry, 0)
	err := do_in_session("entries", func(coll *mgo.Collection) error {
		return coll.Find(bson.M{"readed": false, "category": category}).
			Sort("-created").
			Skip(skip).
			Limit(limit).
			All(rtn)
	})
	return rtn, err
}

func insert_entry(coll *mgo.Collection, entry Entry) error {
	_, err := coll.Upsert(bson.M{"link": entry.Link}, bson.M{"$setOnInsert": &entry})
	return err
}

func (op_entry) SetContent(link, filepath string, words int, imgs []string) error {
	status := ContentStatusFailed
	imgc := len(imgs)
	if len(filepath) > 0 && (words+imgc*128) > 192 {
		status = ContentStatusReady
	}
	cs := ContentStatis{Local: filepath, Words: words, Imgs: imgc}

	return do_in_session("entries", func(coll *mgo.Collection) error {
		return coll.Update(bson.M{"link": link}, bson.M{"$set": bson.M{
			"status":  status,
			"content": cs,
		}, "$push": bson.M{"image": bson.M{"$each": imgs}}})
	})
}

/*
func NewContentStatis(fp string, words, imgs int) ContentStatis {
	c := ContentStatis{Words: words, Imgs: imgs, Local: fp}
	if len(fp) == 0 {
		c.Status = ContentStatusFailed
	} else {
		c.Status = ContentStatusReady
	}
	return c
}
*/
