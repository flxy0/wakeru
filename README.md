# wakeru

Wakeru is a very messily implemented and barebones file uploading and sharing application. It works without a database. It uses generated hashes for "authentification". This approach was inspired by the VPN service Mullvad, who use a customer ID as their sole means of authentification, instead of requiring username and password. Going along with this, wakeru's file storing and serving relies solely on the filesystem. It is thus recommended to use a server with SSD storage to prevent possible bottlenecks under high load. For serving the files only a portion of the hash is displayed so the rest of your uploads isn't accessible to others.

The project has not yet been benchmarked and is, to my knowledge, not in use anywhere. It's also untested... because I am lazy to be quite honest. Baby steps!

I would personally consider anything pre-alpha quality at this point in time. The existing templates are all unstyled and unformatted, while other routes only end in a plaintext response. This will hopefully be addressed in the future™.

---

To add to that, this is my first time actually using Go which means mistakes and other stupidities are very likely. I would be very happy about any sort of feedback on this project and would very much appreciate any sort of pointers on what could be improved!

The idea for this project has been floating around in my head for quite a while and I'm just glad I finally managed to sort of pull it off!