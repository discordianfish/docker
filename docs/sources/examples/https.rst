:title: Docker HTTPS Setup
:description: How to setup docker with https
:keywords: docker, example, https, daemon

.. _running_docker_https:

Running docker with https
=========================

Normally docker runs via http on ``/var/run/docker.sock``

.. code-block:: bash

   sudo docker -d &

If you wish to run docker via https you first need to generate a certificate
and a private key file. How to do this securely is beyond the scope of this
example, however the following command will generate an example one.

Unstrusted, self-signed certificates
------------------------------------

.. code-block:: bash

    openssl genrsa -out server.pem 2048
    openssl req -new -key server.pem -x509 -out server.csr -days 36525



Docker can then run using these certificates. Most commonly you will want to
run docker on a different port that the default unix socket when in https mode.

.. code-block:: bash

    sudo docker -d -tlskey=server.key -tlscert=server.crt -H=tcp://0.0.0.0 -H unix:///var/run/docker.sock

Note that when run in this way, the docker client will not work with docker.

Since this certificate is self-signed, the docker client won't be able to connect.

Full featured CA
----------------

You can use HTTPS with a self-signed certificate by pointing the `tlscacert`
flag to a trusted CA certificate. The easiest way to create such CA and the
server keys, is by using "easy-rsa-2.0". Create a copy somewhere, then
create you CA like this:

.. code-block:: bash

    ./build-ca
    ./build-dh

Now that we have a CA, you can create a server key and certificate to be
used like showed above.

.. code-block:: bash

    ./build-key-server server

Now you can make docker trust the server by pointing `-tlscacert` to your
CA certificate.

.. code-block:: bash

   docker -H=tcp://localhost -tls -tlscacert=/path/to/easy-rsa-2.0/keys/ca.crt

