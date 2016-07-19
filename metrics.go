package hugot

import "github.com/tcolgate/client_golang/prometheus"

var (
	messagesTx = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "hugot_messages_sent_total",
		Help: "Number of messages sent.",
	},
		[]string{"channel"})
	messagesRx = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "hugot_messages_received_total",
		Help: "Number of slack messages received.",
	},
		[]string{"channel"})
)

func init() {
	prometheus.MustRegister(messagesTx)
	prometheus.MustRegister(messagesRx)
}
