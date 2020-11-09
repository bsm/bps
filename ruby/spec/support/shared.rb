require 'securerandom'

RSpec.shared_examples 'publisher' do |features|
  # WARNING: This example group requires the following helpers to be defined by caller:
  #   - `read_messages(topic_name, num_messages)`
  #   - `publisher_url`

  subject do
    BPS::Publisher.resolve(publisher_url)
  end

  after do
    subject.close
  end

  it 'should register' do
    expect(subject).to be_a(described_class)
  end

  it 'should publish' do
    topic_name = "bps-test-topic-#{SecureRandom.uuid}"
    messages = 3.times.map { "bps-test-message-#{SecureRandom.uuid}" }

    topic = subject.topic(topic_name)
    messages.each {|msg| topic.publish(msg) }
    subject.close

    published = read_messages(topic_name, messages.count)
    expect(published).to match_array(messages)
  end
end
