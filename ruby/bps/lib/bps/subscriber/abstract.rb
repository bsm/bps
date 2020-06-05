module BPS
  module Subscriber
    class Abstract
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
