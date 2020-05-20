package kafka

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/bsm/bps"
)

func parseAddrs(u *url.URL) []string {
	if strings.Contains(u.Host, ",") {
		return strings.Split(u.Host, ",")
	}

	addr := u.Host
	if port := u.Port(); port != "" {
		addr += ":" + port
	}
	return []string{addr}
}

func parseCommonQuery(query url.Values) *sarama.Config {
	config := sarama.NewConfig()
	config.ClientID = "bps-client"

	if v := query.Get("client.id"); v != "" {
		config.ClientID = v
	}

	if v := query.Get("kafka.version"); v != "" {
		if kv, err := sarama.ParseKafkaVersion(v); err == nil {
			config.Version = kv
		}
	}

	if v := query.Get("channel.buffer.size"); v != "" {
		config.ChannelBufferSize, _ = strconv.Atoi(v)
	}

	return config
}

func parseProducerQuery(query url.Values) *sarama.Config {
	config := parseCommonQuery(query)

	if v := query.Get("acks"); v != "" {
		if v == "all" {
			config.Producer.RequiredAcks = sarama.WaitForAll
		} else if num, err := strconv.ParseInt(v, 10, 32); err == nil {
			config.Producer.RequiredAcks = sarama.RequiredAcks(num)
		}
	}

	if v := query.Get("message.max.bytes"); v != "" {
		config.Producer.MaxMessageBytes, _ = strconv.Atoi(v)
	}

	switch query.Get("compression.type") {
	case "gzip":
		config.Producer.Compression = sarama.CompressionGZIP
	case "snappy":
		config.Producer.Compression = sarama.CompressionSnappy
	case "lz4":
		config.Producer.Compression = sarama.CompressionLZ4
	}

	switch query.Get("partitioner") {
	case "hash":
		config.Producer.Partitioner = sarama.NewHashPartitioner
	case "random":
		config.Producer.Partitioner = sarama.NewRandomPartitioner
	case "roundrobin":
		config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	}

	if v := query.Get("timeout"); v != "" {
		config.Producer.Timeout, _ = time.ParseDuration(v)
	}

	if v := query.Get("flush.bytes"); v != "" {
		config.Producer.Flush.Bytes, _ = strconv.Atoi(v)
	}
	if v := query.Get("flush.messages"); v != "" {
		config.Producer.Flush.Messages, _ = strconv.Atoi(v)
	}
	if v := query.Get("flush.frequency"); v != "" {
		config.Producer.Flush.Frequency, _ = time.ParseDuration(v)
	}

	if v := query.Get("retry.max"); v != "" {
		config.Producer.Retry.Max, _ = strconv.Atoi(v)
	}
	if v := query.Get("retry.backoff"); v != "" {
		config.Producer.Retry.Backoff, _ = time.ParseDuration(v)
	}

	return config
}

func parseSubscriberQuery(query url.Values) *sarama.Config {
	config := parseCommonQuery(query)

	if v := query.Get("offsets.initial"); v != "" {
		switch v {
		case "newest":
			config.Consumer.Offsets.Initial = sarama.OffsetNewest
		case "oldest":
			config.Consumer.Offsets.Initial = sarama.OffsetOldest
		default:
			config.Consumer.Offsets.Initial, _ = strconv.ParseInt(v, 10, 64)
		}
	}

	// TODO: other defaults seem fine, should we make it fully configurable for initial release?

	return config
}

func convertMessage(topic string, msg *bps.PubMessage) *sarama.ProducerMessage {
	var key sarama.Encoder
	if msg.ID != "" {
		key = sarama.StringEncoder(msg.ID)
	}

	var headers []sarama.RecordHeader
	if n := len(msg.Attributes); n != 0 {
		headers := make([]sarama.RecordHeader, 0, n)
		for key, val := range msg.Attributes {
			headers = append(headers, sarama.RecordHeader{
				Key:   []byte(key),
				Value: []byte(val),
			})
		}
	}

	value := sarama.ByteEncoder(msg.Data)
	return &sarama.ProducerMessage{Topic: topic, Key: key, Value: value, Headers: headers}
}
