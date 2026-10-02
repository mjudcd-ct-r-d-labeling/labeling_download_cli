# MJU Dataset CLI

MJU 라벨링 데이터셋 다운로드를 위한 명령줄 도구입니다. 터미널의 안내에 따라 인증 정보, 다운로드 대상, 저장 경로를 입력하여 영상·입력 로그·라벨 데이터를 내려받을 수 있습니다.

전체 데이터셋 다운로드, 번호별 세션 다운로드, 목록 파일을 이용한 일괄 다운로드를 지원합니다. 중단된 다운로드는 같은 저장 경로에서 이어받을 수 있습니다.

## 목차

- [사용 전 준비](#사용-전-준비)
- [설치](#설치)
- [다운로드](#다운로드)
- [목록 파일 작성](#목록-파일-작성)
- [저장 파일 및 결과 확인](#저장-파일-및-결과-확인)
- [중단 및 이어받기](#중단-및-이어받기)
- [업데이트 및 제거](#업데이트-및-제거)
- [문제 해결](#문제-해결)
- [명령어 참조](#명령어-참조)

## 사용 전 준비

다운로드를 시작하기 전에 다음 항목을 준비합니다.

- **인증 정보**: `User Key`, `Password`, `Token`
- **저장 경로**: 쓰기 권한과 충분한 여유 공간이 있는 폴더의 절대 경로
- **다운로드 대상**: 특정 데이터를 다운로드하는 경우 `classification_number` 또는 번호 목록 파일
- **네트워크 연결**: 인증, 파일 목록 조회 및 다운로드에 필요

`classification_number`는 다운로드 대상을 식별하는 번호입니다. 이 문서에서는 `GC-2024-0001`을 예시로 사용합니다. `session_id`는 해당 번호에 속한 개별 세션의 식별자입니다.

지원 환경은 다음과 같습니다.

| 운영체제 | 지원 아키텍처 | 실행 환경 |
| --- | --- | --- |
| macOS | Apple Silicon (`arm64`), Intel (`amd64`) | 터미널 |
| Linux | `arm64`, `amd64` | 셸 |
| Windows | `amd64` | PowerShell 5.1 이상 |

## 설치

운영체제에 맞는 설치 명령을 실행합니다. 설치 스크립트는 해당 환경에 맞는 최신 공개 버전을 내려받습니다.

### macOS / Linux

```sh
curl -fsSL https://raw.githubusercontent.com/mjudcd-ct-r-d-labeling/labeling_download_cli/main/scripts/install.sh | sh
```

실행 파일은 `/usr/local/bin/mju-dataset`에 설치됩니다. 설치 경로에 쓰기 권한이 없으면 `sudo` 비밀번호를 요청합니다.

### Windows

PowerShell에서 다음 명령을 실행합니다.

```powershell
irm https://raw.githubusercontent.com/mjudcd-ct-r-d-labeling/labeling_download_cli/main/scripts/install.ps1 | iex
```

실행 파일은 `%LOCALAPPDATA%\mju-dataset\mju-dataset.exe`에 설치되며, 설치 폴더가 사용자 PATH에 추가됩니다. 설치 후 명령을 찾지 못하면 PowerShell을 닫고 다시 실행합니다.

### 설치 확인

```sh
mju-dataset --version
```

버전 정보가 출력되면 설치가 완료된 것입니다.

<details>
<summary>특정 버전 설치</summary>

다음 명령의 `2026.05.25.12`를 설치할 버전 태그로 변경합니다.

**macOS / Linux**

```sh
curl -fsSL https://raw.githubusercontent.com/mjudcd-ct-r-d-labeling/labeling_download_cli/main/scripts/install.sh | sh -s -- 2026.05.25.12
```

**Windows PowerShell**

```powershell
& ([scriptblock]::Create((irm 'https://raw.githubusercontent.com/mjudcd-ct-r-d-labeling/labeling_download_cli/main/scripts/install.ps1'))) -Version '2026.05.25.12'
```

</details>

## 다운로드

### 1. 프로그램 실행 및 인증

터미널 또는 PowerShell에서 다음 명령을 실행합니다.

```sh
mju-dataset
```

프로그램이 시작되면 인증 정보를 순서대로 입력하고, 각 입력 후 Enter를 누릅니다.

```text
User Key:
Password:
Token:
```

`Password`와 `Token`의 입력 내용은 화면에 표시되지 않습니다. 인증에 성공하면 `Authenticated.` 메시지와 다운로드 모드 선택 메뉴가 표시됩니다.

### 2. 다운로드 모드 선택

```text
[1] Download all data (existing export)
[2] Download all sessions for one classification number
[3] Download all sessions from a list file
Select mode (1/2/3):
```

목적에 맞는 번호를 입력합니다.

| 선택 | 용도 | 입력 항목 |
| --- | --- | --- |
| `1` 전체 다운로드 | 서버에서 제공하는 전체 내보내기 데이터를 번호별로 다운로드 | 추가 입력 없음 |
| `2` 단일 번호 다운로드 | 한 번호에 속한 모든 다운로드 가능한 세션을 다운로드 | `classification_number` 하나 |
| `3` 목록 기반 다운로드 | 목록에 포함된 여러 번호의 다운로드 가능한 세션을 일괄 다운로드 | 목록 파일의 절대 경로 (파일명·확장자 포함) |

모드 `2`에서는 `Classification number:`에 대상 번호를 입력합니다. 모드 `3`에서는 **파일명과 확장자를 포함한 목록 파일의 절대 경로**를 입력합니다.

```text
List file (absolute path, CSV/XLSX/JSON):
```

예: macOS `/Users/yourname/lists/targets.csv`, Linux `/home/yourname/lists/targets.xlsx`, Windows `C:\Users\yourname\lists\targets.json`.

지정한 파일 하나만 읽습니다. 파일 형식은 [목록 파일 작성](#목록-파일-작성)을 참고하십시오.

번호 기반 모드(`2`, `3`)에서는 영상·입력 로그·라벨 JSONL 세 파일이 모두 저장소에 존재하는 세션만 다운로드합니다. 세션의 승인 상태(`pending`, `approved` 등)는 다운로드 조건에 포함되지 않습니다.

### 3. 다운로드 저장 경로 지정

다음 안내에 다운로드 파일을 저장할 폴더의 절대 경로를 입력합니다.

```text
Download directory (absolute path):
```

| 운영체제 | 경로 예시 |
| --- | --- |
| macOS | `/Users/yourname/mju_dataset` |
| Linux | `/home/yourname/mju_dataset` |
| Windows | `C:\Users\yourname\mju_dataset` |

`yourname`은 실제 사용자 폴더 이름으로 변경합니다. 상대 경로는 사용할 수 없습니다. 폴더가 존재하지 않으면 생성 여부를 확인하며, 생성하려면 `y`를 입력합니다.

목록 기반 다운로드에서는 읽을 목록 파일의 경로와 다운로드 파일을 저장할 폴더를 각각 지정합니다.

### 4. 기존 파일 처리 방식 선택

저장 경로에 다운로드 대상의 완료된 파일이 있으면 다음 메뉴가 표시됩니다.

| 선택 | 처리 방식 | 설명 |
| --- | --- | --- |
| `1` Resume | 이어받기 | 정상 완료된 파일은 건너뛰고, 누락되거나 손상된 파일을 다운로드합니다. 중단된 부분 파일은 이어받습니다. |
| `2` Fresh | 처음부터 다운로드 | 다운로드 대상의 기존 파일을 삭제하고 다시 다운로드합니다. |

중단된 다운로드를 계속하거나 누락된 파일을 보충하려면 `Resume`을 선택합니다. `Fresh`는 기존 파일을 교체하므로 추가 확인 질문에 `y`를 입력한 경우에만 진행됩니다.

### 5. 다운로드 시작 및 진행 확인

서버 조회가 완료되면 다운로드할 게임 수와 파일 수가 표시됩니다. 번호 기반 모드에서는 세션 수도 함께 표시됩니다. `Press Enter to start.` 안내가 나오면 Enter를 눌러 시작합니다.

다운로드 가능한 파일이 없으면 프로그램이 종료됩니다. 번호나 세션이 없거나 필수 파일이 부족한 항목은 다운로드 대상에서 제외되며, [결과 확인](#저장-파일-및-결과-확인)에서 제외 사유를 확인할 수 있습니다.

대화형 터미널의 진행 화면은 다음과 같습니다.

```text
Overall [########------------]  40% 12/30 files  |  18 remaining
Current GC-2024-0001/GC-2024-0001_gameplay.mp4
File    [############--------]  60%  1.2 GiB / 2.0 GiB  |  Downloading
Done 10  Skipped 2  Failed 0  |  Elapsed 2m15s
```

- `Overall`: 전체 파일 기준 진행률과 남은 파일 수
- `Current` / `File`: 현재 파일 경로와 다운로드 진행률
- `Done` / `Skipped` / `Failed`: 완료·건너뜀·실패 건수
- `Elapsed`: 경과 시간

파일 수신이 끝나면 `Verifying file` 상태를 표시한 뒤, 검증을 통과한 파일을 완료로 집계합니다.

출력을 파일로 리다이렉트한 환경에서는 간헐적인 진행 요약을 기록합니다.

## 목록 파일 작성

목록 기반 다운로드(모드 `3`)는 직접 지정한 **CSV·XLSX·JSON 파일 하나**를 읽습니다. 폴더 경로는 사용할 수 없습니다.

아래 형식 중 하나를 사용하여 대상 번호를 작성합니다.

### CSV

UTF-8로 저장하며, 첫 행에 `classification_number` 헤더를 작성합니다. 다른 열은 무시합니다.

```csv
classification_number
GC-2024-0001
GC-2024-0002
```

### XLSX

첫 번째 시트의 첫 행에 `classification_number` 열을 작성하고, 아래 행에 번호를 하나씩 입력합니다. 다른 열과 시트는 무시합니다.

| classification_number |
| --- |
| GC-2024-0001 |
| GC-2024-0002 |

### JSON

`classification_numbers` 키에 번호를 문자열 배열로 작성합니다.

```json
{
  "classification_numbers": [
    "GC-2024-0001",
    "GC-2024-0002"
  ]
}
```

### 파일 처리 기준

- 지정한 파일 안에서 중복된 번호는 한 번만 처리합니다.
- 번호 앞뒤의 공백을 제거하고 빈 값은 무시합니다.
- 각 목록 파일의 최대 크기는 50 MiB입니다.
- 지정한 파일이 없거나, 지원하지 않는 확장자이거나, 번호가 비어 있거나, 파일 형식이 잘못된 경우 오류를 표시합니다.
- 목록 파일 자체는 서버에 업로드하지 않으며, 읽어 들인 번호 배열만 전송합니다.

## 저장 파일 및 결과 확인

### 저장 구조

전체 다운로드(모드 `1`)는 번호별 폴더에 파일을 저장합니다.

```text
<저장 경로>/
└── <classification_number>/
    ├── <classification_number>_gameplay.mp4
    ├── <classification_number>_inputlogs.jsonl
    └── <classification_number>_labeling.jsonl
```

번호 기반 다운로드(모드 `2`, `3`)는 각 번호 아래에 세션별 폴더를 생성합니다.

```text
<저장 경로>/
└── <classification_number>/
    └── <session_id>/
        ├── <classification_number>_gameplay.mp4
        ├── <classification_number>_inputlogs.jsonl
        └── <classification_number>_labeling.jsonl
```

저장 경로에는 다음 설명 파일과 상태 폴더도 생성됩니다.

| 항목 | 설명 |
| --- | --- |
| `data_explain.md` | 데이터셋 설명 문서. 이 문서의 다운로드가 실패해도 데이터 다운로드는 계속됩니다. |
| `.mju-dataset-download/` | 이어받기 상태와 로그를 저장하는 숨김 폴더 |
| `.mju-dataset-download/download.log` | 파일별 다운로드 결과 |
| `.mju-dataset-download/unavailable.txt` | 번호 기반 모드에서 다운로드 대상에서 제외된 항목과 사유. 제외 항목이 있을 때 생성됩니다. |

이어받기가 필요한 경우 부분 다운로드 파일(`.part`)과 `.mju-dataset-download` 폴더를 유지합니다.

### 완료 결과

다운로드가 완료되면 다음 형식으로 결과를 출력합니다.

```text
Done.  Success: <count>  Skipped: <count>  Failed: <count>
```

`Success`는 이번 실행에서 다운로드한 파일 수, `Skipped`는 이미 완료되어 건너뛴 파일 수, `Failed`는 다운로드에 실패한 파일 수입니다. 실패한 파일의 전체 목록은 요약 아래에 표시됩니다.

번호 기반 모드에서 제외된 번호·세션은 화면에 최대 5건을 표시합니다. 전체 제외 목록은 `unavailable.txt`에서 확인할 수 있습니다.

## 중단 및 이어받기

다운로드 중 `Ctrl + C`를 누르면 상태를 보존하고 종료합니다.

```text
Download interrupted. Run again to resume.
```

이어받으려면 다음 순서로 실행합니다.

1. `mju-dataset`을 다시 실행하고 인증 정보를 입력합니다.
2. 이전에 사용한 다운로드 모드와 대상을 선택합니다.
3. 이전과 동일한 다운로드 저장 경로를 입력합니다.
4. 기존 파일 처리 메뉴가 표시되면 `1` (`Resume`)을 선택합니다.
5. 다운로드 대상 수를 확인하고 Enter를 누릅니다.

파일 버전이 변경된 경우에는 이전 부분 파일에 새 내용을 이어 붙이지 않고 다시 다운로드합니다.

## 업데이트 및 제거

### 최신 버전으로 업데이트

실행 중인 CLI를 종료한 뒤 다음 명령을 실행합니다.

```sh
mju-dataset --update
```

데이터셋 인증 없이 최신 공개 버전을 확인하고, 현재 운영체제와 아키텍처에 맞는 실행 파일을 내려받습니다. 파일 크기와 SHA-256 검증을 통과한 경우에만 실행 파일을 교체합니다. 현재 버전이 같거나 더 높으면 교체하지 않습니다.

다운로드한 데이터와 이어받기 상태는 유지됩니다. 업데이트 후 `mju-dataset --version`으로 버전을 확인합니다.

업데이트 파일을 받는 동안 진행률, 전송량, 경과 시간을 표시합니다.

```text
Update [########------------]  40%  8.0 MiB / 20.0 MiB
Status Downloading  |  Elapsed 3s
```

다운로드 후에는 `Verifying size and SHA-256` → `Verified` → `Installing verified update...` 순서로 상태를 안내합니다. `Verified`는 내려받은 파일의 검증 완료를 뜻하며, 설치 완료 여부는 이후 출력되는 메시지로 확인합니다. 검증 또는 저장에 실패하면 `Failed`, 다운로드가 중단되면 `Interrupted`를 표시합니다. 출력을 파일로 리다이렉트하면 화면 갱신 대신 주기적인 진행 요약을 기록합니다.

macOS / Linux에서는 설치 경로에 쓰기 권한이 없으면 `sudo` 비밀번호를 요청합니다. Windows에서는 CLI 종료 후 PowerShell 도우미가 실행 파일을 교체하므로 다른 CLI 인스턴스도 먼저 종료합니다.

`--update`를 지원하지 않는 이전 버전은 [설치 명령](#설치)을 다시 실행합니다.

### CLI 제거

```sh
mju-dataset --uninstall
```

인증 정보 없이 CLI 실행 파일을 제거합니다. 다운로드한 데이터와 이어받기 상태는 유지됩니다.

macOS / Linux에서는 필요하면 `sudo` 비밀번호를 요청합니다. Windows에서는 다른 CLI 인스턴스를 먼저 종료합니다. CLI 종료 후 PowerShell 도우미가 실행 파일과 기본 설치 경로의 사용자 PATH 항목을 정리합니다. 설치 폴더는 비어 있을 때만 제거하며, PATH 변경은 새 터미널에 반영됩니다.

<details>
<summary>이전 버전의 CLI 수동 제거</summary>

`--uninstall`을 지원하지 않는 이전 버전은 기본 설치 경로에서 실행 파일을 직접 제거합니다. 다른 경로에 설치한 경우 실제 설치 경로로 변경합니다.

**macOS / Linux**

```sh
sudo rm -f /usr/local/bin/mju-dataset
```

**Windows PowerShell**

```powershell
Remove-Item "$env:LOCALAPPDATA\mju-dataset\mju-dataset.exe" -Force
```

설치 폴더 전체를 제거하려면 다음 명령을 사용합니다.

```powershell
Remove-Item "$env:LOCALAPPDATA\mju-dataset" -Recurse -Force
```

Windows에서 수동 제거한 경우 사용자 PATH에 남은 설치 경로를 별도로 정리합니다.

</details>

### 다운로드 데이터 제거

다운로드할 때 지정한 저장 폴더를 파일 관리자에서 삭제합니다. 폴더를 삭제하면 데이터 파일, `data_explain.md`, 다운로드 로그 및 이어받기 상태가 함께 삭제됩니다.

## 문제 해결

| 증상 | 조치 |
| --- | --- |
| Windows 설치 후 `mju-dataset` 명령을 찾을 수 없음 | PowerShell을 닫고 다시 실행합니다. |
| 인증 실패 | `User Key`, `Password`, `Token`이 모두 올바른지 확인하고 다시 실행합니다. |
| 저장 경로 오류 | 절대 경로인지 확인합니다. 해당 폴더에 쓰기 권한이 없으면 다른 경로를 지정합니다. |
| 목록 파일 오류 | 헤더·JSON 형식, UTF-8 인코딩(CSV), 파일 크기 및 번호 입력 여부를 확인합니다. |
| 다운로드 대상이 없거나 일부 세션이 제외됨 | 번호와 필수 파일 존재 여부를 확인합니다. 제외 목록이 생성된 경우 `unavailable.txt`의 사유를 확인합니다. |
| 파일 다운로드 실패 | `download.log`와 완료 화면의 실패 목록을 확인한 후 같은 대상·저장 경로로 다시 실행합니다. |
| 업데이트 중 네트워크 오류 또는 GitHub API 요청 제한 | 네트워크 연결을 확인하고 잠시 후 다시 실행합니다. |
| Windows 업데이트 실패 | 설치 폴더의 `.mju-dataset-update-*.log`를 확인합니다. |
| CLI 제거 실패 | 명령에서 안내한 로그 경로를 확인합니다. Windows에서는 임시 로그 경로를 안내합니다. |

`download.log`와 `unavailable.txt`는 다운로드 저장 경로의 `.mju-dataset-download` 폴더에 있습니다.

## 명령어 참조

명령은 터미널 또는 PowerShell에서 실행합니다. 다운로드 대상과 경로는 실행 후 대화형으로 입력합니다.

| 명령어 | 기능 |
| --- | --- |
| `mju-dataset` | 데이터셋 다운로드 시작 |
| `mju-dataset --version` | 설치된 버전 확인 |
| `mju-dataset --update` | 최신 공개 버전으로 업데이트 |
| `mju-dataset --uninstall` | CLI 제거 |

서버 API 계약과 빌드·배포 절차는 [개발자 가이드](docs/development.md)를 참고하십시오.
