require 'bps'
require 'kafka'

module BPS
  module Kafka
    # ReliablePublisher is a slow, but reliable Kafka Publisher.
    # There is no batching at all, messages are delivered one by one.
    class ReliablePublisher < BPS::Publisher::Abstract
      class Topic < BPS::Publisher::Topic::Abstract
        def initialize(producer, topic_name, opts = {})
          @producer = producer
          @opts = opts.merge(topic: topic_name)
        end

        def publish(msg_data)
          @producer.produce(msg_data, **@opts)
          @producer.deliver_messages
          nil
        end
      end
      private_constant :Topic

      def initialize(seed_brokers, opts = {})
        @producer_opts = extract_prefixed_opts!(opts, 'producer_')
        @produce_opts = extract_prefixed_opts!(opts, 'produce_')

        @client = ::Kafka.new(seed_brokers, **opts)
      end

      def topic(name)
        producer = @client.producer(**@producer_opts)
        Topic.new(producer, name, **@produce_opts)
      end

      def close
        @client.close
        nil
      end

      private

      # changes passed opts, returns sub-opts that start with given `prefix` (with `prefix` removed)
      def extract_prefixed_opts!(opts, prefix)
        {}.tap do |extracted|
          opts.delete_if do |name, value|
            next unless name.delete_prefix!(prefix)

            extracted[name] = value
            true
          end
        end
      end
    end
  end

  register_publisher('kafka') do |url, opts|
    addrs = CGI.unescape(url.host).split(',')
    Kafka::ReliablePublisher.new(addrs, **opts)
  end
end
