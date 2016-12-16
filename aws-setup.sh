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


