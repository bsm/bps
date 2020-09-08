module BPS
  module Publisher
    class InMem < Abstract
      class Topic < Abstract::Topic
        attr_reader :name, :messages

        def initialize(name)
          super()
          @name = name.to_s
          @messages = []
        end

        # Publish a message.
        def publish(message, **)
          @messages.push(message.to_s)
        end
      end

      def initialize
        super()
        @topics = {}
      end

      # @return [Array<String>] the existing topic names.
      def topic_names
        @topic_names ||= @topics.keys.sort
      end

      # Retrieve a topic handle.
      # @params [String] name the topic name.
      def topic(name)
        name = name.to_s

        @topics[name] ||= begin
          @topic_names = nil
          self.class::Topic.new(name)
        end
      end
    end
  end
end
