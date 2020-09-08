require 'bps/publisher/kafka'

module BPS
  module Publisher
    class KafkaAsync < Kafka
      class Topic < Kafka::Topic
        private

        def after_publish; end
      end

      # @see BPS::Kafka::Publisher
      #
      # @param [Hash] opts the options.
      # @option opts [Integer] :max_queue_size (defaults to: 1000)
      #                        the maximum number of messages allowed in the queue.
      # @option opts [Integer] :delivery_threshold (defaults to: 0)
      #                        if greater than zero, the number of buffered messages that will automatically
      #                        trigger a delivery.
      # @option opts [Integer] :delivery_interval (defaults to: 0) if greater than zero, the number of
      #                        seconds between automatic message deliveries.
      def initialize(broker_addrs, **opts) # rubocop:disable Lint/UselessMethodDefinition
        super
      end

      private

      def init_producer(max_queue_size: 1000, delivery_threshold: 0, delivery_interval: 0)
        @client.async_producer(
          max_queue_size: max_queue_size,
          delivery_threshold: delivery_threshold,
          delivery_interval: delivery_interval,
        )
      end
    end
  end
end
