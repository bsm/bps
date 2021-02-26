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
      def initialize(cluster_id, client_id, nats: {}, **opts)
        super()

        # handle TLS if CA file is provided:
        if !nats[:tls] && nats[:tls_ca_file]
          ctx = OpenSSL::SSL::SSLContext.new
          ctx.set_params
          ctx.ca_file = nats.delete(:tls_ca_file)
          nats[:tls] = ctx
        end

        @topics = {}
        @client = ::STAN::Client.new
        @client.connect(cluster_id, client_id, nats: nats, **opts.slice(*::BPS::STAN::Utils::CLIENT_OPTS.keys))
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
