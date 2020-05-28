require 'securerandom'

RSpec.shared_examples 'publisher' do |features|
  before do
    expect(features).to have_key(:read_messages) # each publisher test must provide message reader
    expect(features).to have_key(:url)           # each publisher must BPS.register_publisher for at least 1 scheme
  end

  after do
    subject.close
  end

  let(:topic_name) { "bps-test-topic-#{SecureRandom.uuid}" }
  let(:seed_messages) { 3.times.map { "bps-test-message-#{SecureRandom.uuid}" } }

  it 'should publish' do
    topic = subject.topic(topic_name)
    seed_messages.each {|msg| topic.publish(msg) }

    messages = features[:read_messages].call(topic_name)
    expect(messages).to match_array(seed_messages, seed_messages.count)
  end

  it 'should register' do
    Array(features[:url]).each do |url|
      subject = BPS.resolve_publisher(url)
      expect(subject).to be_a(described_class)
      subject.close
    end
  end
end
