package lru

import "container/list"

type Cache struct {
	maxBytes  int64                    //允许使用的最大内存
	nbytes    int64                    //当前已经使用的内存
	ll        *list.List               //双向链表
	cache     map[string]*list.Element //字典存储键和值的映射关系
	OnEvicted func(key string, value Value)
}

//双向链表的数据类型 当淘汰队首节点时 用key从字典中删除相应的映射
type entry struct {
	key   string
	value Value
}

//返回值所占用的内存
type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//如果该节点存在，则将该节点移动到队尾 双向链表为队列，队首队尾是相对的 约定front为队尾
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		//interface类型转换 存储的是任意类型
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *Cache) RemoveOldset() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldset()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
