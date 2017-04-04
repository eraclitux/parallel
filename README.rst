==========
GoParallel
==========

|image0|_ |image1|_

.. |image0| image:: https://godoc.org/github.com/eraclitux/parallel?status.png
.. _image0: https://godoc.org/github.com/eraclitux/parallel

.. |image1| image:: https://travis-ci.org/eraclitux/parallel.svg?branch=master
.. _image1: https://travis-ci.org/eraclitux/parallel

Package ``parallel`` try to simplify use of parallel (as not concurrent) workers that run on their own core.
Number of workers is adjusted at runtime in base of numbers of cores.
This paradigm is particularly useful in presence of heavy, independent tasks.
