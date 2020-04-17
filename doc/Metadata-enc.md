# Metadata encryption

---

#### 1. Overview
- Encryption is currently used to encrypt all cid strings when all the file cids obtained after import are stored in the database library. At this time, the file cids in the database are all encrypted strings and need to be decrypted in a corresponding manner

---
#### 2. encryption
two encryption methods are provided
- aes
- rsa


#### 3. Logic：

The import command can choose whether the pwd parameter needs to be encrypted. There are two methods: aes and rsa


- When obtaining the file cid in the import.go file, the method `importchunk ()

```
func (m *MetaData) importchunk(ctx context.Context, c *chunk, wg *sync.WaitGroup) (string, error) {
    defer wg.Done()

    cid, err := m.StorageAPI.Import(ctx, c.tempfilepath)
    if err != nil {
        return "", xerrors.Errorf("failed to call import api: %w", err)
    }

    //TODO: import data success, delete tempfile
    //Because the source file is retrieved according to cid when placing an order,
    //the source file cannot be deleted here, otherwise the order will fail
    //so can't delete tempfile now

    // if c.absPath != c.tempfilepath {
    //  if err := os.Remove(c.tempfilepath); err != nil {
    //      log.Warnf("delete tempfile failed, fpath: %s", c.tempfilepath)
    //  }
    // }

    cPath := m.chunkPath(c.absPath)
    key := datastore.KeyWithNamespaces([]string{ChunkPrefix, cPath, strconv.Itoa(c.sequence)})

    if m.EncType != "" {
        cid, err = encypt.Encyptdata(m.EncType, cid, AESKEY)
    }

    if err := m.Mstore.DS.Put(key, []byte(cid)); err != nil {
        return "", xerrors.Errorf("saving chunkinfo to store: %w", err)
    }

    return cid, nil
}
```
- When ordering to obtain the cid, determine whether to use encryption and use the corresponding method to decrypt
```
func (d *Deal) flushcidtodb(ctx context.Context, dbname string) error {
    // log.Infof("flushcid to deal : %s", dbname)
    repopath, ok := ctx.Value(repo.CtxRepoPath).(string)
    if !ok {
        return xerrors.New("ctx value repopath not found")
    }

    st, err := NewFSstore(repopath, "meta", dbname)
    if err != nil {
        return err
    }

    //Here to determine whether the cid uses encryption
    var enc string
    e, err := st.DS.Get(datastore.NewKey(EncTypeKey))
    if err != nil {
        enc = ""
    } else {
        enc = string(e)
        PrivateKey, _ = encypt.GetKeysbyType(enc, repopath)
    }

    files, err := st.DS.Query(dsq.Query{Prefix: ChunkPrefix})
    if err != nil {
        return xerrors.Errorf("qurey file err:", err)
    }

    senc, _ := json.Marshal(api.DealStatus{})
    var key string
    for {
        cid, ok := files.NextSync()
        if !ok {
            break
        }
        key = string(cid.Value)
        // cid := string(cids.Value)
        if enc != "" {
            deckey, _ := encypt.Decyptdata(enc, key, PrivateKey)
            key = string(deckey)
        }
        // fmt.Printf("key: %s,  value: %s\n", cid.Key, string(cid.Value))
        d.Dstore.DS.Put(datastore.NewKey(key), senc)
    }

    return nil
}
```
- Similarly, when retrieving files, the obtained cid is also decrypted in the same way；meta/retrive.go `generateFile()`
```
func (ds *MetaStore) generateFile(ctx context.Context, api api.API, miner string, destpath string) error {
    repopath, ok := ctx.Value(repo.CtxRepoPath).(string)
    if !ok {
        return xerrors.New("ctx value repopath not found")
    }

    cs, err := ds.DS.Get(datastore.NewKey(ChunkSizeKey))
    if err != nil {
        return xerrors.Errorf("get chunksize: %w", err)
    }
    chunksize, _ := strconv.ParseUint(string(cs), 10, 64)

    abspath, err := ds.DS.Get(datastore.NewKey(AbsPathKey))
    if err != nil {
        return err
    }

    t, err := ds.DS.Get(datastore.NewKey(TypeKey))
    if err != nil {
        return err
    }
		//Determine whether the encryption is used, and the corresponding decryption method is used
    var enc string
    e, err := ds.DS.Get(datastore.NewKey(EncTypeKey))
    if err != nil {
        enc = ""
    } else {
        enc = string(e)
        PrivateKey, _ = encypt.GetKeysbyType(enc, repopath)
    }

    paths, err := ds.DS.Query(dsq.Query{Prefix: FilePrefix})
    if err != nil {
        return xerrors.Errorf("query fileprefix: %w", err)
    }

    ......

    return nil
}


```
