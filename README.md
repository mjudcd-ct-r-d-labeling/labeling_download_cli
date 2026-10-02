# mju-dataset CLI

MJU 라벨링 데이터셋을 내려받는 전용 CLI입니다.

이 프로그램은 실행 후 대화형으로 인증 정보를 입력받고 다운로드 모드와 로컬 디렉터리를 선택합니다.

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
mju-dataset --update
mju-dataset --uninstall
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

### 3. 다운로드 모드 선택

인증 후 전체 다운로드, 번호 직접 입력, 목록 파일 디렉터리 입력 중 하나를 선택합니다. 번호 직접 입력을 선택하면 `classification_number`를, 목록 파일 디렉터리를 선택하면 CSV·XLSX·JSON 파일이 들어 있는 디렉터리의 절대 경로를 입력합니다. 자세한 목록 파일 형식은 아래의 「classification number로 모든 세션 다운로드」 절에 있습니다.

### 4. 다운로드 경로 입력

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

### 5. 파일 목록 확인

인증과 경로 선택이 끝나면 서버에서 파일 목록을 가져옵니다.

```text
Fetching file list from server... done.
```

다운로드 가능한 파일이 없으면 그대로 종료됩니다.

### 6. Resume 또는 Fresh 선택

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

### 7. 다운로드 시작

준비가 끝나면 게임 수와 파일 수가 표시되고, Enter를 누르면 다운로드가 시작됩니다.

```text
Ready to download X games / Y files.
Press Enter to start.
```

대화형 터미널에서는 아래 네 줄이 갱신됩니다. 파일이 많아도 파일마다 새 줄이 계속 쌓이지 않습니다.

```text
Overall [########------------]  40% 12/30 files  |  18 remaining
Current GC-2024-0001/GC-2024-0001_gameplay.mp4
File    [############--------]  60%  1.2 GiB / 2.0 GiB  |  Downloading
Done 10  Skipped 2  Failed 0  |  Elapsed 2m15s
```

터미널이 아닌 출력(로그 파일 등)에는 ANSI 화면 제어 문자 대신 간헐적인 진행 요약을 남깁니다. 실패한 파일의 전체 목록은 완료 후 출력되고, 파일별 결과는 `.mju-dataset-download/download.log`에도 기록됩니다.

### 8. 중단과 재시작

다운로드 중 `Ctrl + C`로 중단할 수 있습니다.

중단되면 다음과 같이 출력되고 종료됩니다.

```text
Download interrupted. Run again to resume.
```

같은 다운로드 경로로 다시 `mju-dataset`를 실행하면 이어받을 수 있습니다.

### 9. 완료 후 결과 확인

완료되면 아래 형식의 요약이 출력됩니다.

```text
Done.  Success: <count>  Skipped: <count>  Failed: <count>
```

추가로 생성되는 파일과 폴더:

- 선택한 다운로드 경로 아래에 실제 데이터 파일이 저장됩니다.
- `data_explain.md` 파일이 함께 내려받아집니다.
- 숨김 폴더 `.mju-dataset-download`가 생성되며, 여기에는 이어받기 상태와 로그가 저장됩니다.

## 업데이트

CLI를 종료한 상태에서 다음 명령을 실행합니다.

```sh
mju-dataset --update
```

데이터셋 로그인 없이 공개된 배포 태그에서 최신 버전을 확인하고, 현재 OS·CPU에 맞는 바이너리를 내려받습니다. 파일 크기와 SHA-256이 일치할 때만 실행 파일을 교체합니다. 현재 버전이 같거나 더 높으면 교체하지 않습니다. 다운로드한 데이터와 이어받기 상태는 유지됩니다.

- macOS / Linux: 설치 경로의 쓰기 권한이 없으면 `sudo` 비밀번호를 요청합니다.
- Windows: CLI가 종료된 뒤 PowerShell 업데이트 도우미가 실행 파일을 교체합니다. 다른 CLI 인스턴스는 먼저 종료하세요. 실패하면 설치 폴더의 `.mju-dataset-update-*.log`를 확인하세요.
- 업데이트 후 `mju-dataset --version`으로 설치 버전을 확인합니다.
- `--update`가 없는 이전 버전은 위 설치 명령을 한 번 다시 실행해야 합니다.
- GitHub API 요청 제한이나 네트워크 오류가 발생하면 잠시 후 다시 실행합니다.

## 코드 변경과 배포

`main` 브랜치에 push하거나 `main`에서 GitHub Actions를 수동 실행하면 테스트, 5개 플랫폼 빌드, 배포 서버 업로드가 진행됩니다. 모든 플랫폼의 업로드가 성공한 뒤 버전 태그를 공개하므로 설치·업데이트 대상에 포함됩니다. 버전은 커밋 날짜와 Actions 실행 번호로 생성합니다.

필요한 저장소 시크릿은 `DOWNLOADER_API_ENDPOINT`, `API_BASE`, `CLI_RELEASE_PUBLISH_TOKEN`입니다. 마지막 값은 서버의 `CLIENT_RELEASE_PUBLISH_TOKEN`과 일치해야 합니다. 부분 업로드나 실패한 실행은 최신 공개 태그를 갱신하지 않습니다. 실패한 업로드를 재실행할 때 이미 등록된 플랫폼 버전이 충돌하면 서버의 기존 릴리즈를 정리하거나 새 Actions 실행으로 새 버전을 배포해야 합니다.

데이터셋 API 주소와 업데이트용 배포 API 주소는 각각 빌드 시 Base64로 주입합니다. 빌드 후 두 주소가 바이너리에 평문으로 포함되지 않았는지 검사합니다.

로컬 수정·커밋이나 다른 브랜치의 push만으로 배포되지 않습니다. 배포 완료 후에도 사용자는 `mju-dataset --update` 또는 설치 스크립트를 실행해야 합니다.

## 삭제 방법

삭제는 `CLI만 삭제`하는 경우와 `다운로드한 데이터까지 삭제`하는 경우를 구분해서 진행하면 됩니다.

### CLI만 삭제

다음 명령으로 현재 실행한 CLI를 삭제합니다. 인증 정보는 필요하지 않으며, 다운로드한 데이터와 이어받기 상태는 유지됩니다.

```sh
mju-dataset --uninstall
```

macOS / Linux에서는 필요하면 `sudo` 비밀번호를 요청합니다. Windows에서는 CLI 종료 후 PowerShell 도우미가 실행 파일을 삭제하고, 기본 설치 경로의 사용자 PATH 항목을 정리합니다. 기본 설치 폴더는 비어 있을 때만 삭제합니다. 다른 CLI 인스턴스는 먼저 종료하세요. 실패하면 명령에서 안내한 임시 로그를 확인하세요. PATH 변경은 새 터미널에서 반영됩니다.

`--uninstall`이 없는 이전 버전이나 직접 삭제가 필요한 경우에는 아래 명령을 사용합니다.

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

## classification number로 모든 세션 다운로드

인증 후 다음 세 가지 모드 중 하나를 선택합니다. 전체 다운로드 모드의 `/exports/list`, `/exports/file/{classification_number}/{file_type}` API와 저장 방식은 유지합니다.

| 모드 | 입력 | 대상 |
|---|---|---|
| 전체 다운로드 | 다운로드 디렉터리 | 기존 API가 반환하는 번호별 파일 한 세트 |
| 번호 직접 입력 | `classification_number`, 다운로드 디렉터리 | 해당 번호의 모든 세션 |
| 목록 디렉터리 입력 | 목록 파일 디렉터리, 다운로드 디렉터리 | 파일에서 읽은 모든 번호의 모든 세션 |

번호 기반 두 모드에서는 세션 상태(`pending`, `approved` 등)를 다운로드 조건으로 사용하지 않습니다. **영상, 입력 로그, 라벨 JSONL 세 파일이 모두 실제 저장소에 존재하는 세션만 다운로드 대상**입니다. 번호나 세션이 없거나 파일이 부족하면 화면에 최대 5건을 표시하고, 전체 목록은 다운로드 디렉터리의 `.mju-dataset-download/unavailable.txt`에 저장합니다. 다운로드 파일은 다음처럼 세션별 폴더에 저장합니다.

```text
<download-directory>/
└── <classification_number>/
    └── <session_id>/
        ├── <classification_number>_gameplay.mp4
        ├── <classification_number>_inputlogs.jsonl
        └── <classification_number>_labeling.jsonl
```

### 목록 파일 디렉터리와 파일 형식

CLI에 **목록 파일이 들어 있는 디렉터리의 절대 경로**를 입력합니다. 다운로드 파일을 저장할 디렉터리는 별도로 입력합니다. CLI는 지정한 디렉터리의 최상위에 있는 `.csv`, `.xlsx`, `.json` 파일을 모두 읽습니다. 하위 디렉터리는 탐색하지 않으며, 빈 값은 무시하고 앞뒤 공백을 제거한 뒤 중복된 번호는 한 번만 다운로드합니다. 지원 형식의 파일이 없거나 파일 형식이 잘못되면 다운로드 전에 오류를 보여줍니다.

CSV는 UTF-8 텍스트이고 첫 행에 `classification_number` 헤더가 있어야 합니다. 다른 열은 무시합니다.

```csv
classification_number
GC-2024-0001
GC-2024-0002
```

XLSX는 **첫 번째 시트의 첫 행**에 `classification_number` 열을 두고, 아래 행에 번호를 하나씩 입력합니다. 다른 열과 시트는 무시합니다.

JSON은 `classification_numbers` 문자열 배열을 가진 객체 형식입니다.

```json
{
  "classification_numbers": ["GC-2024-0001", "GC-2024-0002"]
}
```

CLI는 파일들을 읽은 뒤 번호·세션·파일 건수를 보여주고 다운로드 시작 여부를 확인합니다. 서버에는 파일을 업로드하지 않고 번호 배열만 전달합니다.

### 서버 API 계약

기존 `User Key`·`Password`·`Token` 인증을 유지합니다. 번호 직접 입력과 목록 디렉터리 입력은 같은 API를 사용합니다.

```http
POST /exports/classifications/manifest
Authorization: Bearer <token>
Content-Type: application/json

{"classification_numbers":["GC-2024-0001","GC-2024-0002"]}
```

요청은 최대 100개 번호씩 전송합니다. 응답은 요청 순서대로 모든 번호와 해당 번호의 모든 세션을 포함합니다. 번호 상태는 `found`, `invalid`, `not_found`, `no_sessions` 중 하나이며, 세션 상태는 `ready` 또는 `incomplete`입니다. 다운로드할 수 없는 번호와 세션에는 `reason`이 포함됩니다. 세션 상태와 무관하게 세 파일이 모두 있어야 `ready`입니다.

```json
{
  "classifications": [{
    "classification_number": "GC-2024-0001",
    "status": "found",
    "sessions": [{
      "session_id": "550e8400-e29b-41d4-a716-446655440000",
      "classification_number": "GC-2024-0001",
      "status": "ready",
      "missing_file_types": [],
      "files": [{
        "file_type": "gameplay",
        "file_name": "GC-2024-0001_gameplay.mp4",
        "download_path": "/exports/sessions/550e8400-e29b-41d4-a716-446655440000/files/gameplay?revision=<revision>",
        "size_bytes": 1048576,
        "sha256": "<64자리 SHA-256>",
        "revision": "<64자리 버전 값>",
        "etag": "\"<revision>\""
      }]
    }]
  }]
}
```

위 예시는 필드 형태를 보여주기 위해 나머지 두 파일을 생략했습니다. 실제 `ready` 세션은 영상·입력 로그·라벨 JSONL 세 항목을 모두 반환합니다. `incomplete` 세션은 `missing_file_types`와 `reason: "required_files_missing"`를 반환하며 다운로드 경로를 제공하지 않습니다. 파일은 아래 경로에서 매 요청마다 인증 및 세션 소속을 검사한 뒤 스트리밍합니다.

```http
GET /exports/sessions/{session_id}/files/{file_type}?revision={revision}
Authorization: Bearer <token>
Range: bytes=1048576-
If-Range: "<etag>"
```

`file_type`은 `gameplay`, `inputlogs`, `labeling`입니다. 이어받기에는 HTTP Range를 사용하고, 파일 버전이 바뀌면 이전 부분 파일에 새 내용을 이어 붙이지 않습니다.

대화형 터미널에는 전체 파일 진행률, 현재 파일의 진행률, 남은 파일 수, 완료·건너뜀·실패 건수가 고정된 화면에 표시됩니다. 출력이 리다이렉트된 환경에는 간헐적인 요약만 기록합니다.

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
