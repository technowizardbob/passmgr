package makeapwd

import (
	"math/rand"
	"time"
)

const pool = "2345689!#%&()=~-^{*}<>?_@[;:],.abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ"

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = pool[rand.Intn(len(pool))]
	}
	return string(bytes)
}

func MakePass(min int, max int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	var n int = randInt(min, max)
	return randomString(n)
}
