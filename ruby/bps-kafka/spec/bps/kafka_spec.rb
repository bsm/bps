def kafka_addrs_str
  ENV.fetch('KAFKA_ADDRS', '127.0.0.1:9092')
end

def kafka_addrs
  kafka_addrs_str.split(',').freeze
end

run_spec = \
  begin
    Kafka.new(kafka_addrs).brokers
    true
  rescue StandardError => e
    warn "WARNING: unable to run #{File.basename __FILE__}: #{e.message}"
    false
  end

RSpec.describe BPS::Kafka::ReliablePublisher, if: run_spec do
  it_behaves_like 'publisher', url: "kafka://#{CGI.escape(kafka_addrs_str)}/" do
    def read_messages(topic_name, num_messages)
      [].tap do |res|
        client = Kafka.new(kafka_addrs)
        # .each_message will block till client is .close-d:
        client.each_message(topic: topic_name, start_from_beginning: true) do |msg|
          res << msg.value
          if res.count >= num_messages
            client.close # the only way to stop consuming
            break
          end
        end
      end
    end
  end
end
