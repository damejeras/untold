# untold

This directory holds encrypted secret keys used in this project.

`untold` allows to embed encrypted secrets into your Go application.
This way you can store encrypted secrets in code repository, and developers can
add new secrets without a need of external secret management tool.

Encrypted passwords are not completely secure. You should never store your passwords
in public repositories, because bad actors can try to decrypt them.
In case of source code leak you should rotate your keys immediately.

Never store private keys in your repository. Private key should be stored someplace secure
and provided to the application within environment variable `UNTOLD_KEY`.
Only exception is your local development environment private key.
