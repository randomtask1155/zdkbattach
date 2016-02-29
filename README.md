# ZDKBATTACH

## Purpose

 Given a KBID and File name this tool will attach the file to a zendesk KB article and delete any existings files with the same name to avoid duplicate file uploads

## Environmental Variables
Required bash variables to run tool
* ZDUSER
* ZDPASS
* ZDROOT

## Usage

```$ zdkbattach
  -f string
    	Provide path of file for upload
  -k string
    	KBID of article you wish to upload file to
```
