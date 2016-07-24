package hugot

import "github.com/prometheus/client_golang/prometheus"

var (
	messagesTx = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "hugot_messages_sent_total",
		Help: "Number of messages sent.",
	},
		[]string{"handler", "channel", "user"})
	messagesRx = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "hugot_messages_received_total",
		Help: "Number of slack messages received.",
	},
		[]string{"handler", "channel", "user"})
)

func init() {
	prometheus.MustRegister(messagesTx)
	prometheus.MustRegister(messagesRx)
}
