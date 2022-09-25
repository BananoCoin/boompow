@echo off

:: Don't change me
set true=1==1
set false=1==0

:: Username (email) If you don't want to be prompted for email on startup
set email=""

:: Password (password) - If you don't want to be prompted for password on startup
set password=""

:: GPU Only - If true, will only compute PoW on GPU, otherwise will use both CPU and GPU
set gpuonly=%false%

:: Max work difficulty
set max_difficulty_multiplier=128

:: Min work difficulty
set min_difficulty_multiplier=1

echo Starting BoomPow Client...

if %gpuonly% (
  boompow-client.exe -email %email% -password %password% -max-difficulty %max_difficulty_multiplier% -min-difficulty %min_difficulty_multiplier% -gpu-only
) else (
  boompow-client.exe -email %email% -password %password% -max-difficulty %max_difficulty_multiplier% -min-difficulty %min_difficulty_multiplier%
)

pause