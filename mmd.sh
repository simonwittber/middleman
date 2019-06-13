screen -AdmS mm -t mmc ./mmd/mmd
sleep 1
screen -S mm -X screen -t auth python3 clients/py/auth.py
sleep 1
screen -S mm -X screen -t chat python3 clients/py/chat.py
screen -r
