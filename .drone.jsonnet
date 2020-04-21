local pipeline = import 'pipeline.libsonnet';

[
  pipeline.test('linux', 'amd64'),

  pipeline.build('docker', 'linux', 'amd64'),
  pipeline.build('docker', 'linux', 'arm64'),
  pipeline.build('docker', 'linux', 'arm'),
  pipeline.notifications('docker', depends_on=[
    'linux-amd64',
    'linux-arm64',
    'linux-arm',
  ]),

  pipeline.build('gcr', 'linux', 'amd64'),
  pipeline.build('gcr', 'linux', 'arm64'),
  pipeline.build('gcr', 'linux', 'arm'),
  pipeline.notifications('gcr', depends_on=[
    'linux-amd64',
    'linux-arm64',
    'linux-arm',
  ]),

  pipeline.build('acr', 'linux', 'amd64'),
  pipeline.build('acr', 'linux', 'arm64'),
  pipeline.build('acr', 'linux', 'arm'),
  pipeline.notifications('acr', depends_on=[
    'linux-amd64',
    'linux-arm64',
    'linux-arm',
  ]),

  pipeline.build('ecr', 'linux', 'amd64'),
  pipeline.build('ecr', 'linux', 'arm64'),
  pipeline.build('ecr', 'linux', 'arm'),
  pipeline.notifications('ecr', depends_on=[
    'linux-amd64',
    'linux-arm64',
    'linux-arm',
  ]),

  pipeline.build('heroku', 'linux', 'amd64'),
  pipeline.build('heroku', 'linux', 'arm64'),
  pipeline.build('heroku', 'linux', 'arm'),
  pipeline.notifications('heroku', depends_on=[
    'linux-amd64',
    'linux-arm64',
    'linux-arm',
  ]),
]
