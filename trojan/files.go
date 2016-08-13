package main

const (
	template string = `upstream {NAME} { server {IP}:{PORT}; }
server {
  listen      [::]:80;
  listen      80;
  server_name {SERVERNAME};

  location    / {
    proxy_pass  http://{NAME};
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $http_host;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Port $server_port;
    proxy_set_header X-Request-Start $msec;

	#{AUTH}
  }
}`

	sslTemplate string = `upstream {NAME} { server {IP}:{PORT}; }
server {
  listen      [::]:80;
  listen      80;
  server_name {SERVERNAME};
  return 301 https://$host$request_uri;
}

server {
  listen      [::]:443 ssl spdy;
  listen      443 ssl spdy;
  server_name {SERVERNAME};

  keepalive_timeout   70;
  add_header          Alternate-Protocol  443:npn-spdy/2;

  location    / {
    proxy_pass  http://{NAME};
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $http_host;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Port $server_port;
    proxy_set_header X-Request-Start $msec;

	#{AUTH}
  }
}`

	nginxConf string = `user www-data;
worker_processes 2;
worker_rlimit_nofile 8192;
pid /run/nginx.pid;

events {
	worker_connections 8000;
	# multi_accept on;
}

http {
	server_tokens off;

	include /etc/nginx/mime.types;
	default_type application/octet-stream;

	# Update charset_types due to updated mime.types
	charset_types text/xml text/plain text/vnd.wap.wml application/x-javascript application/rss+xml text/css application/javascript application/json;

	# Format to use in log files
	log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

	# Logging Settings
	access_log /var/log/nginx/access.log main;
	error_log /var/log/nginx/error.log warn;

	# How long to allow each connection to stay idle; longer values are better
	# for each individual client, particularly for SSL, but means that worker
	# connections are tied up longer. (Default: 65)
	keepalive_timeout 20;

	# Speed up file transfers by using sendfile() to copy directly
	# between descriptors rather than using read()/write().
	sendfile        on;

	# Tell Nginx not to send out partial frames; this increases throughput
	# since TCP frames are filled up before being sent out. (adds TCP_CORK)
	tcp_nopush      on;


	# Compression

	# Enable Gzip compressed.
	gzip on;

	# Compression level (1-9).
	# 5 is a perfect compromise between size and cpu usage, offering about
	# 75% reduction for most ascii files (almost identical to level 9).
	gzip_comp_level    5;

	# Don't compress anything that's already small and unlikely to shrink much
	# if at all (the default is 20 bytes, which is bad as that usually leads to
	# larger files after gzipping).
	gzip_min_length    256;

	# Compress data even for clients that are connecting to us via proxies,
	# identified by the "Via" header (required for CloudFront).
	gzip_proxied       any;

	# Tell proxies to cache both the gzipped and regular version of a resource
	# whenever the client's Accept-Encoding capabilities header varies;
	# Avoids the issue where a non-gzip capable client (which is extremely rare
	# today) would display gibberish if their proxy gave them the gzipped version.
	gzip_vary          on;

	# Compress all output labeled with one of the following MIME-types.
	gzip_types
	application/atom+xml
	application/javascript
	application/json
	application/rss+xml
	application/vnd.ms-fontobject
	application/x-font-ttf
	application/x-web-app-manifest+json
	application/xhtml+xml
	application/xml
	font/opentype
	image/svg+xml
	image/x-icon
	text/css
	text/plain
	text/x-component;
	# text/html is always compressed by HttpGzipModule

    #{InsertSSLHere}
	
	# Additional Configs
	include /etc/nginx/conf.d/*.conf;
}`

	sslOptions string = `
	# SSL
	ssl_session_cache shared:SSL:20m;
	ssl_session_timeout 10m;

	ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-RC4-SHA:ECDHE-RSA-AES128-SHA:AES128-GCM-SHA256:RC4:HIGH:!MD5:!aNULL:!EDH:!CAMELLIA;
	ssl_protocols TLSv1.2 TLSv1.1 TLSv1;
	ssl_prefer_server_ciphers on;
	
	ssl_certificate     {PATH_TO_SERVER_CRT};
  	ssl_certificate_key {PATH_TO_SERVER_KEY};
`
	mimeTypes string = `types {
# Audio
  audio/midi                            mid midi kar;
  audio/mp4                             aac f4a f4b m4a;
  audio/mpeg                            mp3;
  audio/ogg                             oga ogg;
  audio/x-realaudio                     ra;
  audio/x-wav                           wav;

# Images
  image/bmp                             bmp;
  image/gif                             gif;
  image/jpeg                            jpeg jpg;
  image/png                             png;
  image/tiff                            tif tiff;
  image/vnd.wap.wbmp                    wbmp;
  image/webp                            webp;
  image/x-icon                          ico cur;
  image/x-jng                           jng;

# JavaScript
  application/javascript                js;
  application/json                      json;

# Manifest files
  application/x-web-app-manifest+json   webapp;
  text/cache-manifest                   manifest appcache;

# Microsoft Office
  application/msword                                                         doc;
  application/vnd.ms-excel                                                   xls;
  application/vnd.ms-powerpoint                                              ppt;
  application/vnd.openxmlformats-officedocument.wordprocessingml.document    docx;
  application/vnd.openxmlformats-officedocument.spreadsheetml.sheet          xlsx;
  application/vnd.openxmlformats-officedocument.presentationml.presentation  pptx;

# Video
  video/3gpp                            3gpp 3gp;
  video/mp4                             mp4 m4v f4v f4p;
  video/mpeg                            mpeg mpg;
  video/ogg                             ogv;
  video/quicktime                       mov;
  video/webm                            webm;
  video/x-flv                           flv;
  video/x-mng                           mng;
  video/x-ms-asf                        asx asf;
  video/x-ms-wmv                        wmv;
  video/x-msvideo                       avi;

# Web feeds
  application/xml                       atom rdf rss xml;

# Web fonts
  application/font-woff                 woff;
  application/font-woff2                woff2;
  application/vnd.ms-fontobject         eot;
  application/x-font-ttf                ttc ttf;
  font/opentype                         otf;
  image/svg+xml                         svg svgz;

# Other
  application/java-archive              jar war ear;
  application/mac-binhex40              hqx;
  application/pdf                       pdf;
  application/postscript                ps eps ai;
  application/rtf                       rtf;
  application/vnd.wap.wmlc              wmlc;
  application/xhtml+xml                 xhtml;
  application/vnd.google-earth.kml+xml  kml;
  application/vnd.google-earth.kmz      kmz;
  application/x-7z-compressed           7z;
  application/x-chrome-extension        crx;
  application/x-opera-extension         oex;
  application/x-xpinstall               xpi;
  application/x-cocoa                   cco;
  application/x-java-archive-diff       jardiff;
  application/x-java-jnlp-file          jnlp;
  application/x-makeself                run;
  application/x-perl                    pl pm;
  application/x-pilot                   prc pdb;
  application/x-rar-compressed          rar;
  application/x-redhat-package-manager  rpm;
  application/x-sea                     sea;
  application/x-shockwave-flash         swf;
  application/x-stuffit                 sit;
  application/x-tcl                     tcl tk;
  application/x-x509-ca-cert            der pem crt;
  application/x-bittorrent              torrent;
  application/zip                       zip;

  application/octet-stream              bin exe dll;
  application/octet-stream              deb;
  application/octet-stream              dmg;
  application/octet-stream              iso img;
  application/octet-stream              msi msp msm;
  application/octet-stream              safariextz;

  text/css                              css;
  text/html                             html htm shtml;
  text/mathml                           mml;
  text/plain                            txt;
  text/vnd.sun.j2me.app-descriptor      jad;
  text/vnd.wap.wml                      wml;
  text/vtt                              vtt;
  text/x-component                      htc;
  text/x-vcard                          vcf;

}`
)
