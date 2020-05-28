kafka_addrs_str = ENV.fetch('KAFKA_ADDRS', ':9092')
kafka_addrs = kafka_addrs_str.split(',').freeze

run_spec = \
  begin
    Kafka.new(kafka_addrs).brokers
    true
  rescue StandardError => e
    warn "WARNING: unable to run #{File.basename __FILE__}: #{e.message}"
    false
  end

read_messages = proc do |topic_name, num_messages|
  Kafka.new(kafka_addrs).each_message(topic: topic_name).take(num_messages)
end

RSpec.describe BPS::Kafka::ReliablePublisher, if: run_spec do
  subject do
    described_class.new(kafka_addrs)
  end

  it_behaves_like 'publisher',
                  read_messages: read_messages,
                  url: "kafka://#{CGI.escape(kafka_addrs_str)}/"
end
