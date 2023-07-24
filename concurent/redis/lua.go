package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type ScriptSha struct {
	LockSHA,
	UnlockSHA,
	ListPopPushSHA,
	SalesOf7daysSSHA,
	ModifySalesSHA,
	ModifyRanksSHA,
	GetHomeStatic,
	GetSalesInfoSHA string
}

var SHASET ScriptSha

func loadsha(f func() *redis.Script, c context.Context, rdb *redis.ClusterClient) string {
	sha, _ := f().Load(c, rdb).Result()
	return sha
}
func init() {
	rdb := &redis.ClusterClient{}
	c := context.Background()
	SHASET.LockSHA = loadsha(luaLock, c, rdb)
	SHASET.ListPopPushSHA = loadsha(listPopPush, c, rdb)
	SHASET.UnlockSHA = loadsha(luaUnlock, c, rdb)
	SHASET.GetSalesInfoSHA = loadsha(getSalesInfo, c, rdb)
	SHASET.ModifySalesSHA = loadsha(modifySales, c, rdb)
	SHASET.ModifyRanksSHA = loadsha(modifyRanks, c, rdb)
	SHASET.SalesOf7daysSSHA = loadsha(salesOF7days, c, rdb)
	SHASET.GetHomeStatic = loadsha(getHomeStatic, c, rdb)
}

// 加锁lua脚本,设置过期时间 ExLock
func luaLock() *redis.Script {
	return redis.NewScript(`
	local user,dur = ARGV[1],tonumber(ARGV[2])
	local ok= redis.call("set",KEYS[1],user)
	if ok~=nil then
		redis.call("expire",KEYS[1],dur)
		return 1
	end
	return 0
`)
}

// 解锁lua脚本
func luaUnlock() *redis.Script {
	return redis.NewScript(`
	local key = tostring(KEYS[1])
	local userid =tostring(ARGV[1])
	if redis.call("get",key)==ARGV[1] then
		return redis.call("del",key)
	end
`)
}

// 向特定的榜单中添加数据
func listPopPush() *redis.Script {
	return redis.NewScript(`
		local key =KEYS[1]
		local val =tonumber(ARGV[1])
		local lval =tonumber(redis.call("lpop",key))
		if lval ==nil then 
		return 0
		end
		local ll =tonumber(redis.call("lpush",key,lval+val))
		if ll ==nil then
		return 0
		end
		return ll
`)
}

// 获取连续七天销量信息
func salesOF7days() *redis.Script {
	return redis.NewScript(`
	local keys={}
	for i=1,15,1 do
		keys[i]=KEYS[i]
	end
	local array={}
	for i=7,1,-1 do
		local e=redis.call("lindex",keys[16-i],0) or "0"
		array[i]=e
		redis.call("lpush",keys[7],e)
	end
	redis.call("ltrim",keys[7],0,6)
	return array

`)
}

/*
改变销量信息的lua脚本
KEYS[1]今日开始时间,KEYS[2]昨日开始时间,KEYS[3]:日榜，KEYS[4]周榜，KEYS[5]月榜，KEYS[6]年榜，KEYS[7]7天连榜，KEYS[8]总榜
日榜keys[9]-0day,KEYS[10]-1day,keys[11]-2day,KEYS[12]-3day,KEYS[13]-4day,KEYS[14]-5day,KEYS[15]-6day
ARGV[1]订单金额
*/
func modifySales() *redis.Script {

	return redis.NewScript(`
local keys={}
	local val=tonumber(ARGV[1])
	local vals={}
	for i=1,15,1 do
	keys[i]=tostring(KEYS[i])
	end
	if val~=0 then
		for i=3,8,1 do
			vals[i]=tonumber(redis.call("lpop",keys[i])) or 0
			redis.call("lpush",keys[i],vals[i]+val)
		end
	end
	local array={}
	for i=7,1,-1 do
		local e=redis.call("lindex",keys[16-i],0) or "0"
		array[i]=e
		redis.call("lpush",keys[7],e)
	end
	for i=3,7,1 do
	redis.call("ltrim",keys[i],0,6)
	end
	return array
`)
}

// 改变商品排行信息
func modifyRanks() *redis.Script {
	return redis.NewScript(`
	local keys={}
	for i=1,10,1 do
		keys[i]=tostring(KEYS[i])
	end
	local field =tostring(ARGV[1])
	local sales =tonumber(ARGV[2])
	for i=1,10,1 do
		redis.call("zincrby",keys[i],sales,field)
	end
	return 1
	

`)
}

// 商品销量信息
func getSalesInfo() *redis.Script {
	return redis.NewScript(`
	local array={}
	array[5]=redis.call("lindex",KEYS[8],0) or "0"
	array[4]=redis.call("lindex",KEYS[6],0) or "0"
	array[3]=redis.call("lindex",KEYS[5],0) or "0"
	array[2]=redis.call("lindex",KEYS[4],0) or "0"
	array[1]=redis.call("lindex",KEYS[3],0) or "0"
	return array

`)
}

// 获取车主端主页信息
func getHomeStatic() *redis.Script {
	return redis.NewScript(`
	local array={}
	local len =#KEYS
	for i=1,len,1 do
	array[i]=redis.call("lindex",KEYS[i],0) or 0
	end
	return array
`)
}

func Lualock(c context.Context, rdb *redis.ClusterClient, key []string, val ...string) {
	ret := rdb.EvalSha(c, SHASET.LockSHA, []string{"lock"}, val)
	res, err := ret.Result()
	fmt.Println("加锁结果", res, err)
}
func Luaunlock(c context.Context, rdb *redis.ClusterClient, key []string, val ...string) {
	res, err := rdb.EvalSha(c, SHASET.UnlockSHA, []string{"lock"}, val).Result()
	fmt.Println("解锁结果", res, err)
}
