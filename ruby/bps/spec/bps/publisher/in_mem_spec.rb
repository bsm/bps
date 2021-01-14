require 'spec_helper'

RSpec.describe BPS::Publisher::InMem do
  subject { publisher }

  let(:publisher) { described_class.new }

  it 'maintains topics' do
    expect(publisher.topic('x')).to be_a(described_class::Topic)
    expect(publisher.topic_names).to eq(%w[x])
    expect(publisher.topic('z')).to be_a(described_class::Topic)
    expect(publisher.topic_names).to eq(%w[x z])
    expect(publisher.topic('y')).to be_a(described_class::Topic)
    expect(publisher.topic_names).to eq(%w[x y z])
  end

  it 'publishes' do
    topic = publisher.topic('x')
    topic.publish('foo')
    topic.publish('bar')
    expect(topic.messages).to eq(%w[foo bar])

    expect { topic.flush }.not_to raise_error
    expect(publisher.topic('x').messages).to eq(%w[foo bar])
  end
end
