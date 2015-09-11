package ak

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

var sessionIdKey = "SESSIONID"

//定义session
type Session struct {
	sessionId string
	data      map[string]interface{}
	t         time.Time //创建时间
}

//创建session
func newSession() *Session {
	h := md5.New()
	h.Write([]byte(fmt.Sprint("%d", time.Now().Unix())))
	sid := hex.EncodeToString(h.Sum(nil))
	//log.Println("new Session sid = ", sid)
	return &Session{sessionId: sid, data: make(map[string]interface{}), t: time.Now()}
}

//存入session
func (s *Session) Put(k string, v interface{}) {
	s.data[k] = v
}

//获取session数据
func (s *Session) Get(k string) (interface{}, bool) {
	v, ok := s.data[k]
	return v, ok
}

//删除session数据
func (s *Session) Del(k string) {
	delete(s.data, k)
}

//session池
type spool struct {
	pool map[string]*Session
	lock sync.RWMutex
}

//将sesson存入session池
func (s *spool) addSession(session *Session) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pool[session.sessionId] = session
}

//根据sessionId获取session
func (s *spool) getSession(sessionId string) (*Session, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	session, ok := s.pool[sessionId]
	return session, ok
}

//删除session
func (s *spool) removeSession(sessionId string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.pool, sessionId)
}

//创建spool
func newspool() *spool {
	pool := &spool{pool: make(map[string]*Session)}
	go pool.sessionTimeOut()
	return pool
}

//设置session过期 默认15分钟
func (s *spool) sessionTimeOut() {
	tick := time.Tick(15 * time.Minute)
	for {
		now := time.Now()
		select {
		case <-tick:
			for id, sess := range s.pool {
				if now.After(sess.t) {
					s.removeSession(id)
				}
			}
		}
	}
}
