# mju-dataset CLI

MJU 라벨링 데이터셋을 내려받는 전용 CLI입니다.

이 프로그램은 실행 후 대화형으로 인증 정보를 입력받고, 내려받을 로컬 디렉터리를 선택한 뒤 전체 데이터셋을 다운로드합니다.

## 설치

### macOS / Linux

최신 버전 설치:

```sh
curl -fsSL https://raw.githubusercontent.com/mjudcd-ct-r-d-labeling/labeling_download_cli/main/scripts/install.sh | sh
```

특정 버전 설치:

```sh
curl -fsSL https://raw.githubusercontent.com/mjudcd-ct-r-d-labeling/labeling_download_cli/main/scripts/install.sh | sh -s -- 2026.05.25.12
```

설치가 끝나면 버전 확인:

```sh
mju-dataset --version
```

기본 설치 경로:

```text
/usr/local/bin/mju-dataset
```

`/usr/local/bin`에 쓰기 권한이 없으면 설치 중 `sudo` 비밀번호를 묻습니다.

### Windows PowerShell

최신 버전 설치:

```powershell
irm https://raw.githubusercontent.com/mjudcd-ct-r-d-labeling/labeling_download_cli/main/scripts/install.ps1 | iex
```

특정 버전 설치:

```powershell
& ([scriptblock]::Create((irm 'https://raw.githubusercontent.com/mjudcd-ct-r-d-labeling/labeling_download_cli/main/scripts/install.ps1'))) -Version '2026.05.25.12'
```

설치가 끝나면 버전 확인:

```powershell
mju-dataset --version
```

기본 설치 경로:

```text
%LOCALAPPDATA%\mju-dataset\mju-dataset.exe
```

처음 설치 시 사용자 PATH에 설치 경로를 추가합니다. PowerShell을 다시 열어야 명령이 바로 잡힐 수 있습니다.

## 사용 방법

이 CLI는 별도 인자 없이 실행하는 대화형 프로그램입니다.

```sh
mju-dataset
```

지원 옵션:

```sh
mju-dataset --version
```

### 1. 프로그램 실행

터미널에서 아래처럼 실행합니다.

```sh
mju-dataset
```

실행하면 `MJU Labeling Dataset Downloader` 배너가 표시됩니다.

### 2. 인증 정보 입력

아래 3가지를 순서대로 입력합니다.

```text
User Key:
Password:
Token:
```

`Password`와 `Token`은 화면에 표시되지 않습니다.

인증이 성공하면 아래 메시지가 출력됩니다.

```text
Authenticated.
```

### 3. 다운로드 경로 입력

다음 프롬프트가 나오면 절대 경로를 입력합니다.

```text
Download directory (absolute path):
```

예시:

```text
/Users/febook/mju_dataset
```

주의사항:

- 상대 경로는 허용되지 않습니다.
- 디렉터리가 없으면 생성 여부를 묻습니다.
- 쓰기 권한이 없으면 다른 경로를 선택해야 합니다.

### 4. 파일 목록 확인

인증과 경로 선택이 끝나면 서버에서 파일 목록을 가져옵니다.

```text
Fetching file list from server... done.
```

다운로드 가능한 파일이 없으면 그대로 종료됩니다.

### 5. Resume 또는 Fresh 선택

선택한 폴더에 기존 다운로드 데이터가 있으면 아래 메뉴가 나옵니다.

```text
Existing dataset files were found (N verified).
[1] Resume - skip verified files, download missing/corrupted ones
[2] Fresh  - remove/overwrite existing files and download everything again
```

각 선택의 의미:

- `Resume`: 이미 정상 다운로드된 파일은 건너뛰고, 빠졌거나 손상된 파일만 다시 받습니다.
- `Fresh`: 기존 파일을 덮어쓰고 처음부터 다시 받습니다.

`Fresh`를 고르면 한 번 더 확인 질문이 나옵니다.

### 6. 다운로드 시작

준비가 끝나면 게임 수와 파일 수가 표시되고, Enter를 누르면 다운로드가 시작됩니다.

```text
Ready to download X games / Y files.
Press Enter to start.
```

다운로드 중에는 파일별 진행률이 출력됩니다.

### 7. 중단과 재시작

다운로드 중 `Ctrl + C`로 중단할 수 있습니다.

중단되면 다음과 같이 출력되고 종료됩니다.

```text
Download interrupted. Run again to resume.
```

같은 다운로드 경로로 다시 `mju-dataset`를 실행하면 이어받을 수 있습니다.

### 8. 완료 후 결과 확인

완료되면 아래 형식의 요약이 출력됩니다.

```text
Done.  Success: <count>  Skipped: <count>  Failed: <count>
```

추가로 생성되는 파일과 폴더:

- 선택한 다운로드 경로 아래에 실제 데이터 파일이 저장됩니다.
- `data_explain.md` 파일이 함께 내려받아집니다.
- 숨김 폴더 `.mju-dataset-download`가 생성되며, 여기에는 이어받기 상태와 로그가 저장됩니다.

## 삭제 방법

삭제는 `CLI만 삭제`하는 경우와 `다운로드한 데이터까지 삭제`하는 경우를 구분해서 진행하면 됩니다.

### CLI만 삭제

macOS / Linux:

```sh
sudo rm -f /usr/local/bin/mju-dataset
```

Windows PowerShell:

```powershell
Remove-Item "$env:LOCALAPPDATA\mju-dataset\mju-dataset.exe" -Force
```

Windows에서 설치 폴더까지 같이 지우려면:

```powershell
Remove-Item "$env:LOCALAPPDATA\mju-dataset" -Recurse -Force
```

### 다운로드한 데이터까지 삭제

CLI 삭제와는 별개로, 실제 데이터는 사용자가 선택했던 다운로드 폴더에 있습니다.

예:

```sh
rm -rf /Users/febook/mju_dataset
```

이 폴더를 지우면 아래 항목도 함께 삭제됩니다.

- 다운로드한 데이터 파일
- `data_explain.md`
- `.mju-dataset-download` 상태 폴더

## 빠른 요약

설치:

```sh
curl -fsSL https://raw.githubusercontent.com/mjudcd-ct-r-d-labeling/labeling_download_cli/main/scripts/install.sh | sh
```

실행:

```sh
mju-dataset
```

버전 확인:

```sh
mju-dataset --version
```
