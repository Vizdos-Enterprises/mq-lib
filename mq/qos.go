package mq

/*
Ripped *directly* from AMQP091..
https://github.com/rabbitmq/amqp091-go/blob/main/channel.go#L672

Qos controls how many messages or how many bytes the server will try to keep on
the network for consumers before receiving delivery acks.  The intent of Qos is
to make sure the network buffers stay full between the server and client.

With a prefetch count greater than zero, the server will deliver that many
messages to consumers before acknowledgments are received.  The server ignores
this option when consumers are started with noAck because no acknowledgments
are expected or sent.

With a prefetch size greater than zero, the server will try to keep at least
that many bytes of deliveries flushed to the network before receiving
acknowledgments from the consumers.  This option is ignored when consumers are
started with noAck.

When global is true, these Qos settings apply to all existing and future
consumers on all channels on the same connection.  When false, the Channel.Qos
settings will apply to all existing and future consumers on this channel.
*/
type QualityOfService struct {
	PrefetchCount int
	PrefetchSize  int
	Global        bool
}
