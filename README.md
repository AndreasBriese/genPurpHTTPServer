##genPurpHTTPServer - general purpose HTTP webserver

simple, basic secure go/golang http webserver 

===

This is a simple to use HTTP webserver with basic security features. 

It provides security measures mainly through configs given in JSON file 'SERVER.conf'.

- "ServerAddr" (type string): provide server IP and Port to serve from

  + overrides default ServerAddr (localhost:9000) within server.go
  + is overridden by commandline option --address

- "ServerPath" (type string): provide path information for the root directory to be served
  
  + defaults to the server applications directory
  
- "ServerFiles" (type string array): string array with file paths relative to "ServerPath" of all files that should be served by the webserver 

- "AutoTypes" (type string array): string array with file name endings. If the "ServerFiles" array is empty any file within the directory tree below "ServerPath" ending on these endings will be served.

- "RunTarpit" (type bool): any request for a file, that is not on list "ServerFiles" is answered in a spider/penstester confusing way (see https://www.youtube.com/watch?v=I3pNLB3Cq24 )

- "ServerLogging" (type bool): log any request to stdout

if no SERVER.conf is found in the applications root directory, any file below the tree walk will be served on the default "ServerAddr" or configurated by commandline option --address {IP}:{port}

### Compile:

Prerequisite:
You will need go ( http://golang.org ) to be installed on a maschine running with your OS to compile server.go

  1. clone repo from github / cd {downloadPath}/genPurpHTTPServer
  2. go build server.go
  3. edit the JSON file SERVER.conf 
  
copy the server(.exe) together with SERVER.conf to the root directory you want to serve from. 
   
### Start server:

1. open shell
2. change to the root directory with server and SERVER.conf
3. type './server' (or './server -address=":8080"' to change serverAddr for testing
4. (MacOSX: finder double click on 'server' to start with the SERVER.conf configuration)

App server will check through the "ServerFiles" for availability or build a filelist from the "AutoTypes" file endings. The results are shown in the log-output to stdout. 
    
    $ ./server
    2014/09/09 11:58:10 ./SERVER.conf loaded
    2014/09/09 11:58:10 cd -> /Users/andreas_briese/Documents/gocode/src/github.com/AndreasBriese/genPurpHTTPServer/example/
    2014/09/09 11:58:10 Server dir: /Users/andreas_briese/Documents/gocode/src/github.com/AndreasBriese/genPurpHTTPServer/example
    2014/09/09 11:58:10 Server will serve these files:
    2014/09/09 11:58:10    index.html
    2014/09/09 11:58:10    cvs.js
    2014/09/09 11:58:10    fotos/00.jpg
    2014/09/09 11:58:10    fotos/01.jpg
    2014/09/09 11:58:10    fotos/02.jpg
    2014/09/09 11:58:10    fotos/03.jpg
    2014/09/09 11:58:10    fotos/04.jpg
    2014/09/09 11:58:10    fotos/05.jpg
    2014/09/09 11:58:10    fotos/06.jpg
    2014/09/09 11:58:10    fotos/07.jpg
    2014/09/09 11:58:10    fotos/08.jpg
    2014/09/09 11:58:10    fotos/09.jpg
    2014/09/09 11:58:10 Server starts at localhost:9090
    2014/09/09 11:58:10 Spider/pentester tarpit activated!


if you wan't to log to a file (i. e. server.log) type 
  
    $ ./server > server.log
    
or type 
  
    $ nohup ./server &

to demonize server and log to nohup.out

--> feedback's welcome !
