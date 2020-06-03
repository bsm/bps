require 'spec_helper'
require 'kafka'

def kafka_addrs
  ENV.fetch('KAFKA_ADDRS', '127.0.0.1:9092').split(',').freeze
end

run_spec = \
  begin
    ::Kafka.new(kafka_addrs).brokers
    true
  rescue StandardError => e
    warn "WARNING: unable to run #{File.basename __FILE__}: #{e.message}"
    false
  end

helper = proc do
  def read_messages(topic_name, num_messages)
    client = ::Kafka.new(kafka_addrs)
    Enumerator.new do |y|
      client.each_message(topic: topic_name, start_from_beginning: true) do |msg|
        y << msg.value
      end
    end.take(num_messages)
  ensure
    client&.close
  end
end

RSpec.describe 'Kafka', if: run_spec do
  context BPS::Publisher::Kafka do
    it_behaves_like 'publisher', url: "kafka+sync://#{kafka_addrs.join(',')}/", &helper
  end

  context BPS::Publisher::KafkaAsync do
    it_behaves_like 'publisher', url: "kafka://#{kafka_addrs.join(',')}/", &helper
  end
end
