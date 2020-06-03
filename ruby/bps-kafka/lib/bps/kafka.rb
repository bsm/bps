require 'bps'
require 'bps/publisher/kafka'
require 'bps/publisher/kafka_async'

module BPS
  module Publisher
    register('kafka+sync', coercer: Kafka::COERCER) do |url, **opts|
      addrs = CGI.unescape(url.host).split(',')
      Kafka.new(addrs, **opts)
    end

    register('kafka', coercer: Kafka::COERCER) do |url, **opts|
      addrs = CGI.unescape(url.host).split(',')
      KafkaAsync.new(addrs, **opts)
    end
  end
end
