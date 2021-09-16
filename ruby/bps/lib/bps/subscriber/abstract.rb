module BPS
  module Subscriber
    class Abstract
      def initialize
        @uuid = SecureRandom.uuid
        ObjectSpace.define_finalizer(@uuid, proc { close })
      end

      # Subscribe to a topic
      # @params [String] topic the topic name.
      def subscribe(_topic, **)
        raise 'not implemented'
      end

      # Close the subscriber.
      def close
        ObjectSpace.undefine_finalizer(@uuid)
      end
    end
  end
end
