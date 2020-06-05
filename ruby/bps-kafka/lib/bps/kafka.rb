require 'bps'
require 'kafka'
require 'bps/publisher/kafka'
require 'bps/publisher/kafka_async'

module BPS
  module Publisher
    register('kafka+sync') do |url, **opts|
      addrs = CGI.unescape(url.host).split(',')
      Kafka.new(addrs, **Kafka.coercer.coerce(opts))
    end

    register('kafka') do |url, **opts|
      addrs = CGI.unescape(url.host).split(',')
      KafkaAsync.new(addrs, **Kafka.coercer.coerce(opts))
    end
  end
end
