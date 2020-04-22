# filecoin-storage-helper

> An Storage Tool For Filecoin

⚠️ WARNING: Filecoin is under heavy development and breaking changes are highly likely between versions. This library is experimental, It may be broken in part or in entirety.

🧩 Filecoin version: **lotus 0.2.7**



## Table of Contents

- [What is Filecoin StorageHelper](#what is filecoin storagehelper)
- [Install](#Install)
- [Usage](#Usage)

## what is filecoin storagehelper

- Why do we need storage assistants ?

Because currently filecoin does not support files larger than sector size to place storage orders, we design storage assistant to help users cut large files into smaller files and realize the storage requirements of files larger than sector size. In addition, filecoin does not support directory storage, we also implemented the directory storage function, the files in the directory are stored separately to achieve this function.

-  What is metadata ?

Metadata: metadata is used for data retrieval. For example, if the user stores a directory, the metadata information records which files are in the directory (cid of the file); So we think that users have to pay for metadata, that is, they also have to store metadata.

## Install

```shell
# cd {project}

# make all
```



## Usage

To see a full list of commands, run `storagehelper --help`.

expample:

- 1 help
storagehelper --help

- 2 import and get final cid： storagehelper impot
```
need： 
<file>：   file or dir
option：
<vers> :   lotus/fil   default lotus
<size>:    block sizae
<pwd>：  if enc
```

- 3 storagehelper deal
```
need：
<cid> :      file cid
<duration>:	 storage duration
option：
<vers>：  lotus/fil   default lotus
<miner>:  minner id
<askid>： used when vers=fil
<price>:  used when vers=lotus
```

- 4 check state
```
storagehelper state
need：
<cid>： 订单id
option：
<vers>：  lotus/fil   default lotus
```

- 5 retrive
```
storagehelper retrive
need：
<cid>:  file cid
<targetpath>:  
option：
<vers>:  lotus/fil 
<miner>: miner id
```

## Document
[document](/doc/Readme.md)

## License

Dual-licensed under [MIT](https://github.com/arsyun/filecoin-storage-helper/blob/master/LICENSE-MIT) + [Apache 2.0](https://github.com/arsyun/filecoin-storage-helper/blob/master/LICENSE-APACHE)
