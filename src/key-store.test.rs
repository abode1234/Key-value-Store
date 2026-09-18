use crate::*;
use std::env::temp_dir;
use std::fs;

fn unique_path(name: &str) -> String {
    let path = temp_dir().join(format!("kvs_test_{}_{}", name, random_int()));
    path.to_str().unwrap().to_string()
}

#[test]
fn save_data_writes_content_to_file() {
    let path = unique_path("save_data");

    save_data(&path, "hello world \n");

    let content = fs::read_to_string(&path).unwrap();
    assert_eq!(content, "hello world \n");

    fs::remove_file(&path).unwrap();
}

#[test]
fn save_data_overwrites_existing_content() {
    let path = unique_path("save_data_overwrite");

    save_data(&path, "first \n");
    save_data(&path, "second \n");

    let content = fs::read_to_string(&path).unwrap();
    assert_eq!(content, "second \n");

    fs::remove_file(&path).unwrap();
}

#[test]
fn save_data2_writes_content_and_cleans_up_tmp_file() {
    let path = unique_path("save_data2");

    save_data2(&path, "atomic hello \n");

    let content = fs::read_to_string(&path).unwrap();
    assert_eq!(content, "atomic hello \n");

    // The temp file used for the atomic rename should no longer exist.
    let tmp_prefix = format!("{}.tmp.", path);
    let dir = Path::new(&path).parent().unwrap();
    let leftover_tmp = fs::read_dir(dir)
        .unwrap()
        .filter_map(|entry| entry.ok())
        .any(|entry| entry.path().to_str().unwrap().starts_with(&tmp_prefix));
    assert!(!leftover_tmp, "temporary file was not cleaned up after rename");

    fs::remove_file(&path).unwrap();
}

#[test]
fn log_append_appends_lines_in_order() {
    let path = unique_path("log_append");

    let mut fp = log_create(&path);
    log_append(&mut fp, "first line");
    log_append(&mut fp, "second line");
    drop(fp);

    let content = fs::read_to_string(&path).unwrap();
    assert_eq!(content, "first line\nsecond line\n");

    fs::remove_file(&path).unwrap();
}

#[test]
fn log_create_appends_to_existing_file_instead_of_truncating() {
    let path = unique_path("log_create_reopen");

    let mut fp = log_create(&path);
    log_append(&mut fp, "line one");
    drop(fp);

    // Re-open the same log and append more; existing content must survive.
    let mut fp = log_create(&path);
    log_append(&mut fp, "line two");
    drop(fp);

    let content = fs::read_to_string(&path).unwrap();
    assert_eq!(content, "line one\nline two\n");

    fs::remove_file(&path).unwrap();
}
