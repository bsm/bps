require 'bps'
require 'kafka'
require 'bps/publisher/kafka'
require 'bps/publisher/kafka_async'

module BPS
  module Publisher
    register('kafka+sync') do |url, **opts|
      Kafka.new(url, **Kafka.coercer.coerce(opts))
    end

    register('kafka') do |url, **opts|
      KafkaAsync.new(url, **Kafka.coercer.coerce(opts))
    end
  end
end
