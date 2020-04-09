
webpack -p --config webpack.production.config.js
cd production
mv index.bundle.js index_bundle-releasetag.js
go-bindata-assetfs -o bindata_assetfs.go -pkg browser -nocompress=true production/...

mv bindata_assetfs.go browser_assets.go