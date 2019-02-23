package sync

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPool_EmptyPool(t *testing.T) {
	Convey("with empty pool", t, func() {
		pool := NewLimiter(0)

		ticket := takeTicket(pool)

		Convey("no ticket is granted", func() {
			So(ticket, ShouldBeNil)
		})
	})
}

func TestPool_Take(t *testing.T) {
	Convey("with 1 size pool", t, func() {
		pool := NewLimiter(1)

		ticket := takeTicket(pool)

		Convey("ticket is granted", func() {
			So(ticket, ShouldNotBeNil)
		})
	})
}

func TestPool_TakeExhausted(t *testing.T) {
	Convey("with 1 size pool", t, func() {
		pool := NewLimiter(1)

		size1 := pool.Size()
		ticket1 := takeTicket(pool)

		size2 := pool.Size()
		ticket2 := takeTicket(pool)

		pool.Return(ticket1)
		size3 := pool.Size()
		ticket3 := takeTicket(pool)

		Convey("ticket is granted", func() {
			So(size1, ShouldEqual, 1)
			So(ticket1, ShouldNotBeNil)

			So(size2, ShouldEqual, 0)
			So(ticket2, ShouldBeNil)

			So(size3, ShouldEqual, 1)
			So(ticket3, ShouldNotBeNil)
		})
	})
}

func takeTicket(limiter *Limiter) *struct{} {
	select {
	case ticket := <-limiter.Take():
		return ticket
	default:
		return nil
	}
}
