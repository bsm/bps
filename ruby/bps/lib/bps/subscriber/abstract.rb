module BPS
  module Subscriber
    class Abstract
      def initialize
        ObjectSpace.define_finalizer(self, proc { close })
      end

      # Subscribe to a topic
      # @params [String] topic the topic name.
      def subscribe(_topic, **)
        raise 'not implemented'
      end

      # Close the subscriber.
      def close; end
    end
  end
end
