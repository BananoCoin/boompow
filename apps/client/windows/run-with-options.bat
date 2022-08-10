@echo off

:: Don't change me
set true=1==1
set false=1==0

:: Username (email) If you don't want to be prompted for email on startup
set email=""

:: Password (password) - If you don't want to be prompted for password on startup
set password=""

:: Max work difficulty
set max_difficulty_multiplier=128

echo Starting BoomPow Client...

boompow-client.exe -email %email% -password %password% -max-difficulty %max_difficulty_multiplier%

pause