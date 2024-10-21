package ipratelimit

import (
	"time"

	"github.com/Oyal2/tcp-server/pkg/ratelimit"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IpRateLimit", func() {

	var (
		rateLimiter *ratelimit.IPRateLimiter
	)

	const (
		limit    = 5
		interval = time.Second * 2
	)

	BeforeEach(func() {
		var err error
		rateLimiter, err = ratelimit.NewIPRateLimiter(limit, interval)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Allow", func() {
		It("should allow requests up to the limit dealing with a single IP", func() {
			ip := "192.168.1.1"
			for i := 0; i < limit; i++ {
				Expect(rateLimiter.Allow(ip)).To(BeTrue())
			}
			Expect(rateLimiter.Allow(ip)).To(BeFalse())
		})

		It("should track limits separately for each IP dealing with multiple IPs", func() {
			ips := []string{"0.0.0.1", "0.0.0.2", "0.0.0.3"}
			for _, ip := range ips {
				for i := 0; i < limit; i++ {
					Expect(rateLimiter.Allow(ip)).To(BeTrue())
				}
				Expect(rateLimiter.Allow(ip)).To(BeFalse())
			}
		})
	})

	Context("Cleanup", func() {
		It("should remove ips that are passed the interval time", func() {
			ips := []string{"0.0.0.1", "0.0.0.2", "0.0.0.3"}
			for i, ip := range ips {
				Expect(rateLimiter.Allow(ip)).To(BeTrue())
				Expect(rateLimiter.IPs()).NotTo(BeNil())
				Expect(rateLimiter.IPs()).To(HaveLen(i + 1))
				Expect(rateLimiter.IPs()[ip].Count).To(Equal(1))
			}
			//wait for the interval to pass so we can clean those
			time.Sleep(rateLimiter.Interval())

			Expect(rateLimiter.Allow("0.0.0.4")).To(BeTrue())
			rateLimiter.Clean()

			Expect(rateLimiter.IPs()).NotTo(BeNil())
			Expect(rateLimiter.IPs()).To(HaveLen(1))
			Expect(rateLimiter.IPs()["0.0.0.4"].Count).To(Equal(1))
		})
	})
})
