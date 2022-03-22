# wakeru

Wakeru is a very messily implemented and barebones file uploading and sharing application. It works without a database. It uses generated hashes for "authentification". This approach was inspired by the VPN service Mullvad, which uses a customer ID as their sole means of authentification, instead of requiring username and password. Going along with this, wakeru's file storing and serving relies solely on the filesystem. It is thus recommended to use a server with SSD storage or create a large scale instance to prevent possible bottlenecks or other issues through high load. For serving the files only a portion of the hash is displayed so the rest of your uploads isn't accessible to others. Meaning by sharing a file with someone they can't get to your account hash and upload new files to your account or delete your existing ones.

The project has not yet been benchmarked or coded with performance in mind as it is meant for small, personal sharing instances as opposed to something like puush.me or other generalist upload services. It's also untested... because I am lazy to be quite honest. Baby steps!

I would personally consider anything alpha quality at best, so don't rely on it with important data! The existing templates are all lazy styled using [mvp.css](https://andybrewer.github.io/mvp/). My goal isn't to create something beautiful but rather something straight-forward and small. Hence why there is no JavaScript anywhere.

---

To add to that, this is my first Go project that's not just silly experiments or a simple CLI applications. As such, mistakes, oddities and other ~~stupidities~~ silliness are to be expected. I would be very much appreciate any constructive feedback on how things could be improved in any sort of way!

The idea for this project has been floating around in my head for quite a while and I'm just glad I finally managed to bring it to code!

---

TODO:
- [ ] Finish file deletion flow
- [ ] Streamline templates and styling
- [ ] Find solution for custom templates or bake current ones into bin file and make use of a config file for a bit of customisation
- [ ] Block `/generate` URL path but let hashes be generated through another way still (possibly first created hash == admin account which can still generate)