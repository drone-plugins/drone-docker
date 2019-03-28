local pipeline = import 'pipeline.libsonnet';

[
  pipeline.test('windows', 'amd64', '1803'),

  pipeline.build('docker', 'windows', 'amd64', '1803'),
  pipeline.build('docker', 'windows', 'amd64', '1809'),
  pipeline.notifications('docker', 'windows', 'amd64', '1809', [
    'windows-1803',
    'windows-1809'
  ]),

  pipeline.build('gcr', 'windows', 'amd64', '1803'),
  pipeline.build('gcr', 'windows', 'amd64', '1809'),
  pipeline.notifications('gcr', 'windows', 'amd64', '1809', [
    'windows-1803',
    'windows-1809'
  ]),

  pipeline.build('acr', 'windows', 'amd64', '1803'),
  pipeline.build('acr', 'windows', 'amd64', '1809'),
  pipeline.notifications('acr', 'windows', 'amd64', '1809', [
    'windows-1803',
    'windows-1809'
  ]),

  pipeline.build('ecr', 'windows', 'amd64', '1803'),
  pipeline.build('ecr', 'windows', 'amd64', '1809'),
  pipeline.notifications('ecr', 'windows', 'amd64', '1809', [
    'windows-1803',
    'windows-1809'
  ]),
]
