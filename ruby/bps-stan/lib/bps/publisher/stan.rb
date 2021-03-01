require 'cgi'
require 'stan/client'

module BPS
  module Publisher
    class STAN < Abstract
      class Topic < Abstract::Topic
        def initialize(client, topic)
          super()

          @client = client
          @topic = topic
        end

        def publish(message, **_opts)
          @client.publish(@topic, message)
        end

        def flush(**)
          # noop
        end
      end

      # @param [String] cluster ID.
      # @param [String] client ID.
      # @param [Hash] options.
      def initialize(cluster_id, client_id, **opts)
        super()

        @topics = {}
        @client = ::BPS::STAN.connect(cluster_id, client_id, **opts)
      end

      def topic(name)
        @topics[name] ||= self.class::Topic.new(@client, name)
      end

      def close
        # NATS/STAN does not survive multi-closes, so close only once:
        @client&.close
        @client = nil
      end
    end
  end
end
