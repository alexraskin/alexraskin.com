---
title: "PGP"
---

Want to send me something private? Encrypt it to my PGP key and email [hi@alexraskin.com](mailto:hi@alexraskin.com).

**Fingerprint:** `EA38 FEC0 C249 5E0D 685E D874 040F B029 008F D76E`

Download: [pgp.asc](/pgp.asc)

```sh
gpg --keyserver hkps://keys.openpgp.org --recv-keys EA38FEC0C2495E0D685ED874040FB029008FD76E
# or
curl -s https://alexraskin.com/pgp.asc | gpg --import
```

Check that the fingerprint matches before you use it.
