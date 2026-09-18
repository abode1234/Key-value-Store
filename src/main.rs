use std::fs::{File, OpenOptions};
use std::io::Write; // Required to use .write_all() and .sync_all()
use std::path::Path;
use std::time::{SystemTime, UNIX_EPOCH};

pub fn save_data(path: &str, data: &str) {
    // OpenOptions replaces 'w+'
    let mut fp = OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(true)
        .open(path)
        .unwrap();

    fp.write_all(data.as_bytes()).unwrap();
    // Flush to disk (similar to fsync)
    fp.sync_all().unwrap();

    let content = std::fs::read_to_string(path).unwrap();
    println!("{} saved \n", content);
    
    // fp closes automatically here via the Drop trait
    println!("closed \n");
}

fn random_int() -> u128 {
    SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_millis()
}

// Atomic save using rename
pub fn save_data2(path: &str, data: &str) {
    let tmp = format!("{}.tmp.{}", path, random_int());

    let mut fp = OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(true)
        .open(&tmp)
        .unwrap();

    fp.write_all(data.as_bytes()).unwrap();
    fp.sync_all().unwrap();

    let content = std::fs::read_to_string(&tmp).unwrap();
    println!("{} saved \n", content);
    println!("closed \n");

    // Atomic rename replaces the old file safely
    std::fs::rename(&tmp, path).unwrap();
}

// Write-Ahead Log (WAL) create
pub fn log_create(path: &str) -> File {
    OpenOptions::new()
        .create(true)
        .append(true) // Crucial for WAL - matches O_APPEND
        .open(path)
        .unwrap()
}

// Write-Ahead Log (WAL) append
pub fn log_append(fp: &mut File, line: &str) {
    // Convert string to bytes and push the newline
    let mut buf = line.as_bytes().to_vec();
    buf.push(b'\n');
    
    fp.write_all(&buf).unwrap();
    fp.sync_all().unwrap(); 
}

#[cfg(test)]
#[path = "key-store.test.rs"]
mod tests;

fn main() {
    let dir = "data";
    
    // Corrected Path checking
    if !Path::new(dir).exists() {
        std::fs::create_dir(dir).unwrap();
    }
    
    let path = format!("{}/kvs", dir);
    let data = "hello world \n";
    
    println!("without atomic \n");
    
    save_data(&path, &data);
    save_data2(&path, &data);

    let mut fp = log_create(&path);
    log_append(&mut fp, "hello world \n");
}
