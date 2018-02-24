# Video Upload

> Upload videos to Minio so we can transcode them.

The executable should be placed in a directory with two sub-directories of tv and movies. The executable will walk the tree in these directories and upload anything located in them. Once the upload is complete it will delete the local file.

```bash
vids
├── movies
├── tv
└── video-upload.exe
```

## ENV Vars

- MINIO_ACCESS_KEY=\<key\>
- MINIO_SECRET_KEY=\<secret\>
- MINIO_URL=\<minio_url\>
- MINIO_PORT=\<port\>
- MINIO_SECURE=\<true|false\>

## Bucket Notifications

The code will create two buckets

**transcode:** A webhook notification will be added to this bucket to allow for an OpenFaaS function to be called to start a transcoding batch job.

**complete:** A `.keep` file will be added to the complete bucket folders `/tv/` and `/movies`. This is to allow the transcode batch job to have a place to put the transcoded file.

**Note:** I use MakeMKV to rip movies. It supports Windows and my only Blue-ray/DVD drive is in a Windows machine..