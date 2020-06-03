def kafka_addrs_str
  ENV.fetch('KAFKA_ADDRS', '127.0.0.1:9092')
end

def kafka_addrs
  kafka_addrs_str.split(',').freeze
end

def each_kafka_message(client, topic_name)
  Enumerator.new do |y|
    client.each_message(topic: topic_name, start_from_beginning: true) do |msg|
      y << msg.value
    end
  end
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
      client = Kafka.new(kafka_addrs)
      each_kafka_message(client, topic_name).take(num_messages)
    ensure
      client&.close
    end
  end
end
