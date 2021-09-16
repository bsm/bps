require 'bps/kafka'

module BPS
  module Publisher
    class Kafka < Abstract
      class Topic < Abstract::Topic
        def initialize(producer, topic)
          super()

          @producer = producer
          @topic = topic
        end

        def publish(message, **opts)
          @producer.produce(message, **opts, topic: @topic)
          after_publish
        end

        def flush(**)
          @producer.deliver_messages
        end

        protected

        def after_publish
          @producer.deliver_messages
          nil
        end
      end

      CLIENT_OPTS = {
        client_id: :string,
        connect_timeout: :float,
        socket_timeout: :float,
        ssl_ca_cert_file_path: :string,
        ssl_ca_cert: :string,
        ssl_client_cert: :string,
        ssl_client_cert_key: :string,
        ssl_client_cert_key_password: :string,
        ssl_client_cert_chain: :string,
        sasl_gssapi_principal: :string,
        sasl_gssapi_keytab: :string,
        sasl_plain_authzid: :string,
        sasl_plain_username: :string,
        sasl_plain_password: :string,
        sasl_scram_username: :string,
        sasl_scram_password: :string,
        sasl_scram_mechanism: :string,
        sasl_over_ssl: :bool,
        ssl_ca_certs_from_system: :bool,
        ssl_verify_hostname: :bool,
      }.freeze

      PRODUCER_OPTS = {
        # standard
        retry_backoff: :float,
        compression_codec: :symbol,
        compression_threshold: :int,
        ack_timeout: :float,
        required_acks: :symbol,
        max_retries: :int,
        max_buffer_size: :int,
        max_buffer_bytesize: :int,
        idempotent: :bool,
        transactional: :bool,
        transactional_id: :string,
        transactional_timeout: :bool,
        # async
        delivery_interval: :float,
        delivery_threshold: :int,
        max_queue_size: :int,
      }.freeze

      # @return [BPS::Coercer] the options coercer.
      def self.coercer
        @coercer ||= BPS::Coercer.new(CLIENT_OPTS.merge(PRODUCER_OPTS)).freeze
      end

      # @param [Array<String>,URI] brokers the seed broker addresses.
      # @param [Hash] opts the options.
      # @see https://www.rubydoc.info/gems/ruby-kafka/Kafka/Client#initialize-instance_method
      def initialize(broker_addrs, **opts)
        super()

        broker_addrs = parse_url(broker_addrs) if broker_addrs.is_a?(URI)
        @topics = {}
        @client = ::Kafka.new(broker_addrs, **opts.slice(*CLIENT_OPTS.keys))
        @producer = init_producer(**opts.slice(*PRODUCER_OPTS.keys))
      end

      def topic(name)
        @topics[name] ||= self.class::Topic.new(@producer, name)
      end

      def close
        super

        @producer.shutdown
        @client.close
      end

      private

      def parse_url(url)
        port = url.port&.to_s || '9092'
        CGI.unescape(url.host).split(',').map do |addr|
          addr << ':' << port unless /:\d+$/.match?(addr)
          addr
        end
      end

      def init_producer(**opts)
        @producer = @client.producer(**opts)
      end
    end
  end
end
