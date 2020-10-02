require 'spec_helper'
require 'bps/kafka'

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

RSpec.describe 'Kafka' do
  context 'resolve addrs' do
    let(:client) { double('Kafka', producer: nil) }
    before       { allow(::Kafka).to receive(:new).and_return(client) }

    it 'should resolve simple URLs' do
      if Gem::Version.new(RUBY_VERSION) < Gem::Version.new('2.7')
        expect(::Kafka).to receive(:new).with(['test.host:9092'], {}).and_return(client)
      else
        expect(::Kafka).to receive(:new).with(['test.host:9092']).and_return(client)
      end
      BPS::Publisher.resolve(URI.parse('kafka+sync://test.host:9092'))
    end

    it 'should resolve URLs with multiple hosts' do
      if Gem::Version.new(RUBY_VERSION) < Gem::Version.new('2.7')
        expect(::Kafka).to receive(:new).with(['foo.host:9092', 'bar.host:9092'], {}).and_return(client)
      else
        expect(::Kafka).to receive(:new).with(['foo.host:9092', 'bar.host:9092']).and_return(client)
      end
      BPS::Publisher.resolve(URI.parse('kafka+sync://foo.host,bar.host:9092'))
    end

    it 'should resolve URLs with multiple hosts/ports' do
      if Gem::Version.new(RUBY_VERSION) < Gem::Version.new('2.7')
        expect(::Kafka).to receive(:new).with(['foo.host:9093', 'bar.host:9092'], {}).and_return(client)
      else
        expect(::Kafka).to receive(:new).with(['foo.host:9093', 'bar.host:9092']).and_return(client)
      end
      BPS::Publisher.resolve(URI.parse('kafka+sync://foo.host%3A9093,bar.host'))
    end
  end

  context BPS::Publisher::Kafka, if: run_spec do
    it_behaves_like 'publisher', url: "kafka+sync://#{CGI.escape(kafka_addrs.join(','))}/", &helper
  end

  context BPS::Publisher::KafkaAsync, if: run_spec do
    it_behaves_like 'publisher', url: "kafka://#{CGI.escape(kafka_addrs.join(','))}/", &helper
  end
end
