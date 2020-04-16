# filecoin-storage-helper

> An Storage Tool For Filecoin

⚠️ WARNING: Filecoin is under heavy development and breaking changes are highly likely between versions. This library is experimental, It may be broken in part or in entirety.

🧩 Filecoin version: **lotus 0.2.7**



## Table of Contents

- [What is Filecoin StorageHelper](#what is filecoin storagehelper)
- [Install](#Insatll)
- [Usage](#Usage)

## what is filecoin storagehelper

- Why do we need storage assistants ?

Because currently filecoin does not support files larger than sector size to place storage orders, we design storage assistant to help users cut large files into smaller files and realize the storage requirements of files larger than sector size. In addition, filecoin does not support directory storage, we also implemented the directory storage function, the files in the directory are stored separately to achieve this function.

-  What is metadata ?

Metadata: metadata is used for data retrieval. For example, if the user stores a directory, the metadata information records which files are in the directory (cid of the file); So we think that users have to pay for metadata, that is, they also have to store metadata.

## Insatll

```shell
# cd {project}

# make all
```



## Usage

To see a full list of commands, run `storagehelper --help`.

expample:

- import file:

```shell
# storagehelper import  <filepath> 
```
