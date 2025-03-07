# 概览

<img src="https://github.com/skylar2826/go-cache/blob/main/readme.assets/image-20250307164203619.png" />

## 锁 sync.Mutex

<img src="https://github.com/skylar2826/go-cache/blob/main/readme.assets/mutex.png"  width="400px" />

竞争锁的两种模式：

- 正常模式：和新来的（未释放资源，占着processor）一起抢锁，大概率失败（保证效率）
- 饥饿模式：肯定能拿到锁（退出饥饿模式：队列中只剩下一个goroutine或者等待时间小于1ms

## channel

### chansend

<img src="https://github.com/skylar2826/go-cache/blob/main/readme.assets/chansend.png"  width="500px" />

### chanrecv

<img src="https://github.com/skylar2826/go-cache/blob/main/readme.assets/chanrecv.png"  width="500px" />

## 缓存

```
type Cache interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Get(ctx context.Context, key string) (any, error)
	Delete(ctx context.Context, key string) error
}
```

### 过期时间如何控制？

**策略一：每个key开一个goroutine盯着**

缺点：goroutine数量随着key增大；大部分goroutine大多数时间被阻塞，每个都盯着浪费资源；

**策略二：定时轮询**：goroutinue定时轮询所有key

注意事项：定时轮询间隔时间过短，消耗资源；key数量超大时，需要限制，比如限制轮询key的个数或者轮询时长

问题：定时轮询不能访问到所有key，所以有过期key没处理，不能单独定时轮询。可搭配Get时判断一起使用。

**策略三：Get时判断**（懒惰删除）：取数据时再判断是否已过期

### 缓存满了怎么办？

- 控制缓存使用：限制缓存占用内存量、限制缓存键值对

- 使用lru算法淘汰缓存

### 缓存策略

- cache-aside
- read-through
- write-though
- write-back
- refresh-ahead

### 缓存异常

- 穿透：大量不存在的key访问redis不存在，打到db上

- 击穿：key不在缓存中，但在db中

  一般情况下，击穿不会有问题，访问一次db就能回写redis；但供给者短时间大量访问这个key，可能会压垮数据库

- 雪崩：大量key同时过期

### 如何解决缓存异常？

所有问题的落脚点：大量请求达到db上

#### singleflight

大量goroutine同时访问一个key时，singleflight会让其他的原地等待，只有一个访问db

#### 缓存穿透解决方案

1. 使用singleflight
2. 知道数据库不存在key，缓存未命中就返回
   1. 缓存里面是全量数据 =》 未命中就返回
   2. 使用布隆过滤器、bit array等结构，未命中先问依稀啊这些结构
3. 缓存没有，不查db，直接返回默认值
4. 回查db时添加限流器

#### 缓存雪崩解决方案

过期时间添加随机数偏移量

## 分布式锁

分布式锁就是不同实例在网络上抢一把锁，普通锁就是不同线程间在本地竞争锁

使用redis setnx能够排他的设置一个键值对，适合做分布式锁

### 锁需要设置过期时间，原因

不设置过期时间，容易死锁。比如持有锁的实例宕机就死锁了

#### 过期时间应该设置多久？

过短容易过期，过长浪费资源。而且理论上，极限情况下，无论锁时间设置多长，都可能任务没有执行完就过期。

所以在业务推测一个合理的过期时间外，我们还要有续约机制。

### 锁需要保证唯一性，原因

比如：当释放锁时，要释放的是自己锁住的，不能把别人锁着的释放掉

场景举例：任务task1加锁后执行，锁过期释放；task2竞争得到锁，task2加锁；task1执行完成后，过来想释放锁需要比对这把锁是不是自己配置的，不能把task2的加锁释放掉



