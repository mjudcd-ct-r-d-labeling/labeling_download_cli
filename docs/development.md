# 개발자 가이드

이 문서는 mju-dataset CLI의 서버 API 계약과 배포 절차를 설명합니다. 설치 및 다운로드 방법은 [사용자 설명서](../README.md)를 참고하십시오.

## 서버 API 계약

전체 다운로드는 `/exports/list`와 `/exports/file/{classification_number}/{file_type}`을 사용하며, 파일을 번호별 폴더에 저장합니다.

기존 `User Key`·`Password`·`Token` 인증을 유지합니다. 번호 직접 입력과 목록 파일 경로 입력은 같은 API를 사용합니다.

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

## 코드 변경과 배포

`main` 브랜치에 push하거나 `main`에서 GitHub Actions를 수동 실행하면 테스트, 5개 플랫폼 빌드, 배포 서버 업로드가 진행됩니다. 모든 플랫폼의 업로드가 성공한 뒤 버전 태그를 공개하므로 설치·업데이트 대상에 포함됩니다. 버전은 커밋 날짜와 Actions 실행 번호로 생성합니다.

필요한 저장소 시크릿은 `DOWNLOADER_API_ENDPOINT`, `API_BASE`, `CLI_RELEASE_PUBLISH_TOKEN`입니다. 마지막 값은 서버의 `CLIENT_RELEASE_PUBLISH_TOKEN`과 일치해야 합니다. 부분 업로드나 실패한 실행은 최신 공개 태그를 갱신하지 않습니다. 실패한 업로드를 재실행할 때 이미 등록된 플랫폼 버전이 충돌하면 서버의 기존 릴리즈를 정리하거나 새 Actions 실행으로 새 버전을 배포해야 합니다.

데이터셋 API 주소와 업데이트용 배포 API 주소는 각각 빌드 시 Base64로 주입합니다. 빌드 후 두 주소가 바이너리에 평문으로 포함되지 않았는지 검사합니다.

로컬 수정·커밋이나 다른 브랜치의 push만으로 배포되지 않습니다. 배포 완료 후에도 사용자는 `mju-dataset --update` 또는 설치 스크립트를 실행해야 합니다.
