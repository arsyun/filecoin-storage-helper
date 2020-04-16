# go-filecoin-storage-helper

> An Storage Tool For Filecoin

### Storage workflow

We think the keypoint is that client must pay for the metadata, the metadata is used for data retrieval,We'll store the metadata by leveldb, Each upload action will use a separate db.


![](/picture/storage-helper-workflow.jpg)

- Step1
  - Input: User data
  - Output: DB of user data's metadata 
  - Workflow:
    - Record the directory tree to the db(leveldb) through storage helper, including the chunk size, type(the metadata if from user data or db), file path/size.
    - import or make a deal on filecoin
    - get the result from go-filecoin, including the CID of piece and the deal status
    - Record the CID of piece and deal status into db, record  the CID relationship to the original file(file that belongs to,and chunk's seq)
- Step2
  - Input: DB of userdata's metadata
  - Output: DB of DB's metadata
  - Workflow:
    - if the DB of userdata's metadata is great than a sector size,  like as step 1, Think of the DB as a file. split the DB to many chunks.
- Step 3
  - Repeat the step2 until the DB's size is less than a sector size.then import the DB, the CID is the finally result.



## Retrieval workflow

- step1
  - Get the DB through a CID
- step2
  - Get the userdata's metadata that including file path and their CID
- step3
  - Get the file chunk through its CID
- step4
  - Compose the original user data through the file chunks
