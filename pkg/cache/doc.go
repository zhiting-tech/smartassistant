// cache 是一个可以使用go-cache或redis作为缓存存储的缓存库。
// 默认使用基于内存的go-cache(patrickmn/go-cache)作为缓存存储,
// 缓存过期时间为5分钟，可以通过InitCache方法修改默认缓存存储。
// 提供了Set,SetNX,Get,GetWithTLL,Delete这些常用操作缓存的方法。
package cache
