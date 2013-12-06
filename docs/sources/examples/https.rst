:title: Docker HTTPS Setup
:description: How to setup docker with https
:keywords: docker, example, https, daemon

.. _running_docker_https:

Running docker with https
=========================

Normally, docker runs via a non-networked unix socket. It can also optionally
communicate using a http socket.

If you need docker reachable via the network in a safe manner, you can enable
TLS by pointing docker's `tlscacert` flag to trusted CA certificate.

In daemon mode, it will only allow connections from clients authenticates via a
certificate signed by that CA. In client mode, it will only connect to servers
with a certificate signed by that CA.

A easy way to create such CA, server and client keys, is by using
`easy-rsa-2.0`_.
Create a copy somewhere, then create your CA like this:

.. _easy-rsa-2.0: https://github.com/OpenVPN/easy-rsa/tree/release/2.x/easy-rsa/2.0

.. code-block:: bash

    $ ./build-ca
    $ ./build-dh

Now that we have a CA, you can create a server key and certificate. Make sure
that the common name matches the hostname you will use to connect to docker or
just use '*' for a certificate valid for any hostname:

.. code-block:: bash

    $ ./build-key-server server

For client authentication, create a client key and certificate:

.. code-block:: bash

    $ ./build-key client

Now you can make docker daemon only accept connections from clients providing
a certificate trusted by our CA:

.. code-block:: bash

    $ sudo docker -d -tlscacert=ca.crt -tlscert=server.crt -tlskey=server.key -H=tcp://0.0.0.0

To be able to connect to docker, you now need to provide your client keys and
certificates:

.. code-block:: bash

   $ docker -tlscacert=ca.crt -tlscert=client.crt -tlskey=client.key -H=tcp://0.0.0.0

.. warning::

  As shown in the example above, you don't have to run the client
  ``docker`` with ``sudo`` or the ``docker`` group when you use
  certificate authentication. That means anyone with the keys can
  give any instructions to your Docker daemon, giving them root
  access to the machine hosting the daemon. Guard these keys as you
  would a root password!

Other modes
-----------
If you don't want to have complete two-way authentication, you can run docker in
various other modes by mixing the flags.

Daemon modes
~~~~~~~~~~~~
- tlscacert, tlscert, tlskey set: Authenticate clients
- tlscert, tlskey: Do not authenticate clients

Client modes
~~~~~~~~~~~~
- tls: Authenticate server based on public/default CA pool
- tlscacert: Authenticate server based on given CA
- tlscert, tlskey: Authenticate with client certificate, do not authenticate
  server based on given CA
- tlscacert, tlscert, tlskey: Authenticate with client certificate, authenticate
  server based on given CA

