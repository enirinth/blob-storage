# Add ssh pub key
echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDNDMpuhCUyYIQ0kRxAe40/ZNefNxXsgO19uDkztk7CagXRqP80fi9QE2KuF2yjoiAC3FecYbYcjPF57xfMEhE4rwVUiXZtio6oPMWoUb7Qx6oh5Cvg2HUk9kKEi1KmcrYNzPA7QjGdTCOO31hy4b/GVbsiWgYpd0SZw7Rcjk1ZUxVAqSazAMULeSbLKzT6XKfNN/KRYkOfbqRGmcYZ/KoNpS6Gcj8d047Zg1mwVeSxIJ1Pg9kYupndOIGs+fxXwuK72ROM753Z3NDZwI7xnIQtLIoQHgnmPGOnAFIMvZbDIR7U1LCS8+T9rNFH/Y4nTFlz2ZMl6T+zc+OLQlpW6BTtiyW78y1YBhfKWYaij6zs5yhzEahcCitcQPLB9hjtGL6mltSwj2Gc/K3If29dpXbVFZYS5AtcgYVlr0Wn+4Z69q3pzeXgVt2ipfAZOtMzAkOOb/TlW6oZHw8kiCp6HuoBayiKj1mrrPJe59ynG7EiuugLDm8heaULQ41BUEBcuplscOpXOiIPxBk51K8qEgFFSrQouUtO9mtmzu2IOCB4JljNcvXwVZcLglUV53f+tSJhmsxtirtHttOSVNbMOND1A6TSzucUJNNSD6WNNdYRTA4DNvkbw9ERPWX1Rdv3+GZnaDHMflSFFh9BNaJeemiWNq1oy3YZJyJTWMWQ8QVBlw== enirinth@gmail.com" >> ~/.ssh/authorized_keys
echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCzh+oigXPq9PdwhLeCtS2zPuV7TFc7tUXFMuHwD3ndr2uTUKwSYSFMRWhrOQsFb5yeZaGuS0WFK+N0vkm6lzq1fzXwJcOuFrGthAH6oqQCQu/Tb6FKfktb9s3bqjoWJSmIxPauKIiG5SN6zHrDWUeomVS2hi8voaU91vcj9K+sIq0s/mLRZJQlJ6kx/Dhc7KIlmT2qyNgJSlzTIAqhjCZSoOI6SDC95YNCyXqLHmCkgj2n6zzVuI4sJdzZGIl5QrfX+NBmX78+RM3EE9nwL2nquzqPk3CtGXeAVZG6TFD1cGU1lahBgXFEzhtPG7XFEoKrskWJ+F+KBOjdVJcPIKu5 Arjun@Arjuns-MBP.westell.com" >> ~/.ssh/authorized_keys
echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC4XarWxgthi5HUuWbkcKfsaaV0Ho/EzDOpV5vcTfViEudgimdfwEg/k83pSm7+j3LF6IYehtqeIgTi5aSdeZ4VDFQArQQGEby4oLq5taFKrit/mJCLC880y1Y1ybkUtwzVEhjFmw6nPhxfRSVd1OaosTqo+MWZBHjN1IIBaBs33bflwMP17ZJ3Rn7AkdWtHrVC9sVEMajV8et1MRbX5NF4GqwnY3qujlEtu8HWeFGlsocoaeNDteXCbsp3f6NsdxzvUyGJemPfs+cfZQ5EFmxTRpf2R3uYFDuO5VYbSqxRxPIZXwVEO7XiytIATiq2nmY02zjB99GoTMo9Oe1tx8mZDOQh/d22oat3/3rtEkYtuxs4heB2lfX0AueiK140RGhcdNgzssKfEi0B+LzPPeEjARsiuzADsZbUENwaOkIWFYdfg/BZtfuNGUSH5g8Vu1ehRP2y2tJVp90irad/Olomr81DRyZZf+vYFvdrUqeNlq2+MIuMlMQq0qspFKz4hNL4k0xqJRF/+FUPduGN5LFCMDvhggffH8+DV8wFgIdeDWa/BiuMmvHhCmONGGE8ya57lA9sM1TG1NEma8iJkXQysDRByLbnkg0UJCPg0Pd+Kn+KPIk9uqVPqMQddCBtxbt3nxNt0x9CM0V6e4x7qLYetnOnZkHi0me+haIHvALtYQ== raychen0411@gmail.com" >> ~/.ssh/authorized_keys

# Install dependencies
sudo yum -y update
sudo yum -y install git
sudo yum -y install ncurses-devel
sudo yum -y groupinstall 'Development Tools'  # gcc/g++, make, rpm, etc.
git clone https://github.com/vim/vim.git && cd vim/src && sudo make install && cd 
sudo yum -y install java-1.8.0-openjdk-devel

# Install Go
sudo yum -y install wget
sudo wget https://storage.googleapis.com/golang/go1.7.linux-amd64.tar.gz
sudo tar -xvf go1.7.linux-amd64.tar.gz
sudo mv go /usr/local

# make sudo work with go command
sudo ln -s /usr/local/go/bin/go /usr/bin/go 

# Go path setttings
mkdir -p $HOME/workspace/gowork

export GOROOT=/usr/local/go
export GOPATH=$HOME/workspace/gowork
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

echo "export GOROOT=/usr/local/go" >> ~/.bashrc
echo "export GOPATH=\$HOME/workspace/gowork" >> ~/.bashrc
echo "export PATH=\$GOPATH/bin:\$GOROOT/bin:\$PATH" >> ~/.bashrc

# Get project source code (and dependencies)
go get -u github.com/Sirupsen/logrus
go get -u github.com/tatsushid/go-fastping
go get -u github.com/enirinth/blob-storage

# Useful utilities
echo "alias cd582='cd \$HOME/workspace/gowork/src/github.com/enirinth/blob-storage'" >> ~/.bashrc

# Remember to source ~/.bashrc after running this script


