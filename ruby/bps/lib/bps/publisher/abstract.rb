module BPS
  module Publisher
    class Abstract
      class Topic
        # Publish a message.
        def publish(_message, **_opts)
          raise 'not implemented'
        end

        # Flush any remaining buffer.
        def flush(**); end
      end

      def initialize
        ObjectSpace.define_finalizer(self, proc { close })
      end

      # Retrieve a topic handle.
      # @params [String] name the topic name.
      def topic(_name)
        raise 'not implemented'
      end

      # Close the publisher.
      def close; end
    end
  end
end
